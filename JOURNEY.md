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
| ch7 | 7.2 멀티 노드풀 | ⬜ | | |
| ch7 | 7.3 App of Apps | ⬜ | | |
| ch7 | 7.4 멀티테넌시 | ⬜ | | |
| ch8 | 8.1 메시징 | ⬜ | | |
| ch8 | 8.2 트레이싱 | ⬜ | | |
| ch8 | 8.3 CronJob | ⬜ | | |
| ch9 | 9.1 저장소 분석 | ⬜ | | |
| ch9 | 9.2 회고 | ⬜ | | |
| ch9 | 9.3 온보딩 문서 | ⬜ | | |
| ch9 | 9.4 GitAIOps 분석 | ⬜ | | |
| ch9 | 9.5 마무리 | ⬜ | | |

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

## 현재 버전

| 컴포넌트 | 버전 | 변경 이력 |
|---------|------|----------|
| Go | `1.25.0` | 최초 ch2.6에서 Go 1.25 표준 탑재 |
| Notiflex 이미지 | `asia-northeast3-docker.pkg.dev/claude-study-501117/notiflex/api:v0.5.0` | ch6.3 최종 Canary 승격 완료본 반영 |
| ArgoCD | `quay.io/argoproj/argocd:v3.4.5` | ch3.2 최초 설치 및 ch6.2 재구축 적용 |
| Kafka | `(미설치)` | ch8.1 도입 예정 |
| OTel SDK | `(미설치)` | ch8.2 도입 예정 |

## 현재 리소스

| 노드풀 | 머신 타입 | 노드 수 | 주요 워크로드 |
|--------|----------|---------|-------------|
| default-pool | `e2-medium` | `2` | API Pod, Valkey, ArgoCD, Prometheus, Loki, Fluent Bit |

## 트러블슈팅 이력

독자가 겪은 문제와 해결 방법을 기록한다. 같은 문제를 다시 겪지 않도록 한다.

| 챕터 | 문제 | 해결 |
|------|------|------|
| ch6.2 | Secret Manager API 미활성화 상태로 인한 IAM SA 생성 시 무한 대기 현상 | `gcloud services enable secretmanager.googleapis.com` 명령으로 수동 활성화 후 재생성 |
| ch6.2 | Valkey standalone 및 CSI DaemonSet 설치 후 노드 vCPU requests 예산 초과(Pending) | `kube-prometheus-stack` (Operator, Prometheus, Grafana, Alertmanager), Loki, Fluent Bit 의 CPU requests를 모두 `5m`으로 낮추어 리소스 최적화 |
