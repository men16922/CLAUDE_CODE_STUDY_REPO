# Notiflex 여정 기록

이 파일은 독자가 실제로 진행한 내용을 기록한다. AI가 각 챕터 완료 시 자동으로 업데이트한다.

## 진행 현황

| 챕터 | 서브챕터 | 상태 | 완료일 | 비고 |
|------|---------|------|--------|------|
| ch2 | 2.2 설치 확인 | ✅ | 2026-07-16 | |
| ch2 | 2.3 gcloud 설정 | ✅ | 2026-07-16 | |
| ch2 | 2.4 GitHub 저장소 | ✅ | 2026-07-16 | |
| ch2 | 2.5 GKE 클러스터 | ✅ | 2026-07-16 | |
| ch2 | 2.6 빌드/배포 | ✅ | 2026-07-16 | |
| ch2 | 2.7 첫 커밋 | ✅ | 2026-07-16 | |
| ch3 | 3.2 GitOps 도구 | ✅ | 2026-07-16 | ArgoCD 적용 |
| ch3 | 3.3 기능 추가 | ✅ | 2026-07-16 | |
| ch3 | 3.4 CI | ✅ | 2026-07-16 | GitHub Actions 적용 |
| ch3 | 3.5 CI-CD 연결 | ✅ | 2026-07-16 | |
| ch4 | 4.2 메트릭 모니터링 | ✅ | 2026-07-16 | Prometheus + Grafana 구성 |
| ch4 | 4.3 로그 수집 | ✅ | 2026-07-16 | Loki + Fluent Bit 구성 |
| ch4 | 4.4 알림 | ✅ | 2026-07-16 | PrometheusRule 구성 |
| ch5 | 5.2 트래픽 관리 | ✅ | 2026-07-16 | Gateway API 적용 |
| ch5 | 5.3 무중단 배포 | ✅ | 2026-07-16 | Argo Rollouts Blue/Green 구성 |
| ch6 | 6.1 캐시 | ✅ | 2026-07-16 | Valkey standalone 구성 |
| ch6 | 6.2 시크릿 관리 | ✅ | 2026-07-16 | GSM CSI Driver + WI 적용 |
| ch6 | 6.3 Canary 전환 | ✅ | 2026-07-16 | Canary 점진 전환 적용 |
| ch7 | 7.2 멀티 노드풀 | ✅ | 2026-07-25 | nodeSelector 기반 api-pool, worker-pool, ops-pool 설계 |
| ch7 | 7.3 App of Apps | ✅ | 2026-07-25 | root-app 및 Sync Wave 1, 2 적용 |
| ch7 | 7.4 멀티테넌시 | ✅ | 2026-07-25 | enterprise namespace 및 Cross-Namespace DNS 적용 |
| ch8 | 8.1 메시징 | ✅ | 2026-07-25 | Strimzi Operator 기반 KRaft Kafka 구성 |
| ch8 | 8.2 트레이싱 | ✅ | 2026-07-25 | Grafana Tempo 및 OpenTelemetry Go SDK 구성 |
| ch8 | 8.3 CronJob | ✅ | 2026-07-25 | K8s CronJob 기반 5분 주기 헬스체크 구성 |
| ch9 | 9.1 저장소 분석 | ✅ | 2026-07-25 | 저장소 구조, Git 커밋 및 클러스터 상태 통합 분석 |
| ch9 | 9.2 회고 | ✅ | 2026-07-25 | 아키텍처 의사결정 회고 및 CLAUDE.md 성장 분석 |
| ch9 | 9.3 온보딩 문서 | ✅ | 2026-07-25 | docs/onboarding.md 살아있는 온보딩 가이드 생성 |
| ch9 | 9.4 GitAIOps 분석 | ✅ | 2026-07-25 | Git + AI + Ops 삼각 구도 및 루프 패턴 정립 |
| ch9 | 9.5 마무리 | ✅ | 2026-07-25 | 프로덕션 전환 제언 및 GitAIOps 인프라 최종 정비 |

## 도구 선택 기록

독자가 3-프롬프트 패턴(탐색→비교→실행)에서 실제로 선택한 도구와 이유를 기록한다.

| 영역 | 선택 | 검토한 대안 | 선택 이유 |
|------|------|-----------|----------|
| GitOps 도구 (ch3.2) | ArgoCD | Flux | 선언적 Git 동기화 보장 및 웹 콘솔 기반 실시간 시각적 모니터링 우수 |
| CI 도구 (ch3.4) | GitHub Actions | Cloud Build, GitLab CI | 비공개 GitHub 저장소와의 무설정 연동 및 YAML 파이프라인 관리 편의성 |
| 메트릭 (ch4.2) | Prometheus + Grafana | Datadog, New Relic | 오픈소스 시계열 모니터링 표준화 및 차트 커스텀 튜닝 가능성 |
| 로깅 (ch4.3) | Loki + Fluent Bit | ELK Stack | 가볍고 인덱싱 비용이 적으며 Grafana 대시보드와 완벽 호환 |
| 알림 (ch4.4) | PrometheusRule | - | K8s 네이티브 CRD 설정을 통한 알림 룰의 GitOps화 가능 |
| 트래픽 관리 (ch5.2) | Gateway API | NGINX Ingress | GKE L7 외부 로드밸런서 활용으로 인프라 리소스를 절약하고 weight 라우팅 연동 지원 |
| 무중단 배포 (ch5.3) | Argo Rollouts B/G | K8s Deployment | preview 환경 선검증 및 문제 발생 시 즉각 롤백(undo)의 안정성 확보 |
| 캐시 (ch6.1) | Valkey | Redis, Memcached | Redis 프로토콜 및 원자적 카운터(`INCR`) 완벽 호환과 라이선스 우려 배제 |
| 시크릿 관리 (ch6.2) | GSM CSI + WI | K8s Secret, Vault | 암호 키 탈취가 불가능한 무키(keyless) 구조 및 Google Cloud 감사 로그 지원 |
| Canary 배포 (ch6.3) | Argo Rollouts Canary | Blue/Green | 배포 중 순간 리소스 2배 소요 한계 극복 및 단계적 사용자 통제 가능 |
| 노드 배치 (ch7.2) | nodeSelector | Taint/Toleration, Node Affinity | 가장 단순하며 GKE 자동 라벨(`cloud.google.com/gke-nodepool`)을 즉시 활용 가능 |
| 다수 앱 관리 (ch7.3) | App of Apps + Sync Wave | ApplicationSet, 수동 관리 | root-app 하나로 여러 Application의 설치 순서(Sync Wave)와 라이프사이클을 안전하게 관리 |
| 멀티테넌시 (ch7.4) | Namespace + RBAC | vCluster, 물리 클러스터 분리 | 추가적인 연산 비용 없이 논리적 리소스 격리와 Cross-Namespace DNS 호환 제공 |
| 메시징 (ch8.1) | Kafka + Strimzi | RabbitMQ, NATS, Redis Streams | 이벤트 스트리밍 표준, KRaft 모드 지원, Strimzi CRD로 ArgoCD GitOps 완벽 호환 |
| 트레이싱 (ch8.2) | Grafana Tempo | Jaeger, Zipkin | Grafana 내장 연동, OpenTelemetry(OTLP gRPC) 표준 호환, 경량 단일 바이너리 제공 |
| 배치 자동화 (ch8.3) | K8s CronJob | Argo Workflows, Airflow | 별도 추가 설치 없이 K8s 기본 기능 활용, YAML 기반 GitOps 관리 및 자동 력 정제 지원 |

## 현재 버전

| 컴포넌트 | 버전 | 변경 이력 |
|---------|------|----------|
| Go | `1.25.0` | 최초 ch2.6에서 Go 1.25 표준 탑재 |
| Notiflex 이미지 | `asia-northeast3-docker.pkg.dev/claude-study-501117/notiflex/api:v0.7.0` | ch8.2 OpenTelemetry & Kafka 연동 완료본 반영 |
| ArgoCD | `quay.io/argoproj/argocd:v3.4.5` | ch3.2 최초 설치 및 ch6.2 재구축 적용 |
| Kafka | `4.1.0` (Strimzi) | ch8.1 KRaft 단일 브로커 모드 구축 |
| OTel SDK | OpenTelemetry Go SDK | ch8.2 OTLP gRPC 기반 Grafana Tempo 연동 |

## 현재 리소스

| 노드풀 | 머신 타입 | 노드 수 | 주요 워크로드 |
|--------|----------|---------|-------------|
| default-pool | `e2-medium` | 2 | 시스템 컴포넌트 및 관측 가능성 스택 |
| api-pool | `e2-medium` | 1 | Notiflex API (SMB x 2, Enterprise x 1), Valkey |
| worker-pool | `e2-standard-2` | 1 | Strimzi Kafka Operator & Kafka KRaft Broker |
| ops-pool | `e2-small` | 1 | Grafana Tempo & Notiflex Healthcheck CronJob |

## 트러블슈팅 이력

독자가 겪은 문제와 해결 방법을 기록한다. 같은 문제를 다시 겪지 않도록 한다.

| 챕터 | 문제 | 해결 |
|------|------|------|
| ch6.2 | Secret Manager API 미활성화 상태로 인한 IAM SA 생성 시 무한 대기 현상 | `gcloud services enable secretmanager.googleapis.com` 명령으로 수동 활성화 후 재생성 |
| ch6.2 | Valkey standalone 및 CSI DaemonSet 설치 후 노드 vCPU requests 예산 초과(Pending) | `kube-prometheus-stack` (Operator, Prometheus, Grafana, Alertmanager), Loki, Fluent Bit 의 CPU requests를 모두 `5m`으로 낮추어 리소스 최적화 |
