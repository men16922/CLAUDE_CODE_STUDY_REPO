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
