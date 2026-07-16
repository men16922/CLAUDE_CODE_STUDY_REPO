# Notiflex GitOps 실습 여정 기록

## 2장. 개발 환경 구성과 첫 배포

### 1. GCP 환경 설정
- **프로젝트 ID**: `claude-study-501117` (전역 설정을 변경하지 않고 `.env` 파일과 명령어 파라미터로 제어)
- **리전**: `asia-northeast3` (서울)
- **존**: `asia-northeast3-a`

### 2. GKE 클러스터 생성
- **이름**: `notiflex-cluster`
- **타입**: Standard Zonal
- **스펙**: `e2-medium` 노드 2대, Spot VM 적용
- **특이사항**: `--gateway-api=standard` 옵션을 주어 클러스터 생성

### 3. Notiflex API 서버 구성 및 배포
- **코드**: Go 1.25 표준 라이브러리(`net/http`) 기반 상태 확인(`/health`) 및 고유 ID 생성(`/id`) API 구현
- **빌드**: `golang:1.25-alpine`에서 빌드하여 `scratch`로 실행하는 멀티스테이지 Dockerfile 구성
- **레지스트리**: Artifact Registry `notiflex` 리포지토리에 `v0.1.0` 이미지 빌드 및 푸시
- **배포**: Kubernetes namespace (`notiflex`), deployment (`replicas: 2`), service (`ClusterIP`) 리소스 반영 및 포트포워딩을 통한 API 동작 검증 완료

## 3장. 첫 번째 배포 파이프라인

### 1. Argo CD 설치 및 GitOps 연동
- **설치**: GKE 클러스터에 Argo CD v2.14.11 설치 및 Pod 정상 동작 확인
- **비공개 저장소 연동**: GitHub Access Token (`gh auth token`)을 활용하여 repository Secret을 생성해 Argo CD와 GitHub 비공개 저장소 연동 완료
- **Application 생성**: `notiflex-smb` Application을 생성하여 `k8s/smb` 디렉토리 감시 설정 (`syncPolicy`: `automated` with `prune`, `selfHeal`)

### 2. 롤링 업데이트 및 GitOps 배포 검증
- **기능 추가**: `app/main.go`에 `/version` 엔드포인트(v0.1.1) 추가
- **수동 빌드 및 배포**: Cloud Build로 `v0.1.1` 이미지 빌드 후, 매니페스트 이미지 태그 수정 및 `git push`를 통해 Argo CD의 무중단 롤링 업데이트 롤아웃 및 파드 정상 교체 확인

### 3. GitHub Actions CI/CD 파이프라인 구축
- **GCP 서비스 계정 설정**: `github-ci` 서비스 계정 생성 및 `roles/artifactregistry.writer` 권한 연동 완료
- **GitHub Secrets 등록**: `GCP_PROJECT_ID` 및 `GCP_SA_KEY` (JSON 키) 등록 완료
- **워크플로우 구현**: `.github/workflows/ci.yaml`을 추가하여 `app/**` 푸시 시 자동으로 Docker 이미지를 빌드해 Artifact Registry에 푸시하고, `deployment.yaml` 내의 이미지 태그를 커밋 SHA로 변경한 후 커밋 및 푸시하도록 구현
- **전체 파이프라인 검증**: 코드 수정 및 푸시 시, GitHub Actions 빌드 -> `deployment.yaml` 이미지 태그 갱신 및 Git push -> Argo CD가 이를 감지하여 GKE 클러스터에 롤아웃하는 통합 파이프라인 동작 확인 완료 (최종 배포 이미지 태그: `7eac328`)

## 4장. 관측 가능성 한 번에 구축하기

### 1. Prometheus + Grafana 구성 (메트릭 모니터링)
- **설치**: `kube-prometheus-stack` Helm chart를 사용해 메트릭 수집 및 시각화 환경을 구축했습니다.
- **리소스 튜닝**: e2-medium 노드(2대) 자원 한계를 고려하여 `helm-values/kube-prometheus.yaml`을 생성하고 CPU/메모리 Requests를 최소한으로 제한(`Prometheus 100m`, `Grafana 50m`, `Alertmanager 25m`)해 배포했습니다.
- **Loki 데이터 소스 연동**: Grafana 프로비저닝 설정(`additionalDataSources`)에 Loki 서비스를 지정하여, Grafana UI 내에서 Loki 로그를 즉시 쿼리할 수 있도록 구성했습니다.

### 2. Loki + Fluent Bit 구성 (중앙 로그 수집)
- **Loki 설치**: 최소 리소스(10m/128Mi) 및 SingleBinary 모드로 Loki(`grafana/loki`)를 배포 완료했습니다.
- **Fluent Bit 연동**: DaemonSet 형태로 각 노드에 배포되는 `grafana/fluent-bit`를 설치하고, output host를 `loki`로 설정하여 모든 컨테이너의 stdout 로그가 Loki로 안전하게 전송되도록 구축했습니다.
- **검증**: `notiflex-api` 애플리케이션의 롤링 재시작을 통해 부팅 로그(`Starting server on port :8080`)가 정상 생성되도록 하고, Grafana Explore 뷰에서 `{namespace="notiflex"}` 쿼리를 통해 로그 수집 화면을 검증 완료했습니다.

### 3. PrometheusRule 설정 (임계값 알림)
- **리소스 정의**: `k8s/monitoring/pod-restart-alert.yaml`에 `notiflex` namespace 내의 Pod가 5분 내 2회 이상 재시작되는 것을 감지하는 `PodRestartTooMany` 경고 규칙을 생성하여 클러스터에 반영 완료했습니다.

### 4. [추가 실험] Platform Agent (Kiro) 아키텍처 수립
- **아이디어**: 자연어 명령과 단일 ServiceSpec YAML 조합으로 멀티클라우드(AWS, GCP, Azure, On-Prem)에 동일한 파이프라인(Build ➔ Push ➔ Deploy ➔ Validate)을 태우는 AI 플랫폼 에이전트(Kiro) 아키텍처를 정의하고 이를 `Chapter4.md`에 추가 실험으로 문서화했습니다.



## 5장. 무중단 배포

### 1. Gateway API 도입 (외부 트래픽 관리)
- **상태**: ✅ (완료일: 2026-07-16)
- **도구 선택**: GKE Gateway API (`gke-l7-regional-external-managed`)
- **이유**: NGINX Ingress Controller 등 자체 컨트롤러 구축 리소스를 아끼고 GKE 네이티브 기능을 활용해 weight 조절을 수월하게 하기 위함
- **결과**: 외부 IP `35.216.7.224`를 통해 `/health`, `/id`, `/version` 정상 확인 및 `HealthCheckPolicy`(/health:8080) 연동 완료

### 2. Blue/Green 배포 도입 (Argo Rollouts)
- **상태**: ✅ (완료일: 2026-07-16)
- **도구 선택**: Argo Rollouts (`BlueGreen` 전략)
- **이유**: 배포와 검증의 단계를 완벽히 분리하고 preview 환경에서 선검증 후 auto-promote(30초) 전환하기 위함
- **결과**: `notiflex-api` Deployment를 Rollout 리소스로 대체하고 `v0.2.0` 점진 전환 및 롤백용 ReplicaSet 딜레이 유지 구조 수립 완료
- **CI 연동**: `.github/workflows/ci.yaml` 내 sed/add 대상을 `rollout.yaml`로 이전 완료

### 3. 아키텍처 결정 기록(ADR) 작성
- **상태**: ✅ (완료일: 2026-07-16)
- **결과**: `docs/architecture-decisions.md` 신설 및 ADR-001 ~ ADR-007 작성 완료

## 6장. 엔터프라이즈를 위한 기반 정비

### 1. Valkey 캐시 구성 (Pod 간 상태 공유)
- **상태**: ✅ (완료일: 2026-07-16)
- **도구 선택**: Valkey (`standalone` 모드)
- **이유**: Redis 호환 프로토콜과 원자적 증가(`INCR`) 연동을 통해 여러 Pod가 동작하는 환경에서의 분산 카운터 상태 공유(ID 중복 해결) 및 라이선스 제약 극복
- **결과**: `valkey-primary-0` 기동 및 Go app 내 백오프 연결 재시도(10회) 구현 완료. 외부 게이트웨이 요청을 통해 ID가 중복 없이 1, 2, 3... 순차적으로 단일 카운팅되는 흐름 검증 완료. 자원 절약을 위해 replicas를 2에서 1로 선제 축소.

### 2. Google Secret Manager 연동 (시크릿 보안 관리)
- **상태**: ✅ (완료일: 2026-07-16)
- **도구 선택**: GKE Secret Manager CSI Driver (`secrets-store-gke.csi.k8s.io`) 및 Workload Identity
- **이유**: K8s Secret의 평문(Base64) 한계를 넘고, GCP SA 키 마운트가 필요 없는 무키(keyless) 구조로 외부 시크릿 단일 진실 공급원(SSOT) 및 접근 감사를 확보하기 위함
- **결과**: `valkey-password` 시크릿 업로드 및 GKE CSI 애드온 활성화 완료. K8s default ServiceAccount 와 GCP IAM SA 간 Workload Identity 바인딩 ➔ Pod에 `/mnt/secrets/valkey-password` 볼륨 파일 마운트 연동 성공 및 Go app 내 파일 기반 password 로딩 구조(v0.4.0) 검증 완료.

### 3. Canary 배포 전환 (점진적 배포 관리)
- **상태**: ✅ (완료일: 2026-07-16)
- **도구 선택**: Argo Rollouts (`Canary` 전략: 20% ➔ 50% ➔ 80% ➔ 100%, 각 30초 대기)
- **이유**: Blue/Green 배포 시의 2배 리소스 과소비 문제를 극복하고 단계적 트래픽 노출을 통한 세밀한 장애 통제를 위함
- **결과**: `strategy.canary`로 전환 및 구 버전 리소스 클린업 후 `v0.5.0` 배포를 통해 20% -> 50% -> 80% -> 100% 점진적 트래픽 롤아웃 현황 검증 완료.

### 4. 아키텍처 스냅샷 정리
- **상태**: ✅ (완료일: 2026-07-16)
- **결과**: `claude-context/architecture.md` 신설을 통한 AI용 1페이지 아키텍처 토폴로지 구축 및 `docs/architecture-decisions.md`에 ADR-008~010 추가 완료
