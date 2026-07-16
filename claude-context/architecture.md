# Notiflex Platform Architecture Snapshot

이 문서는 Notiflex SaaS 플랫폼의 현재 시점 아키텍처 스냅샷을 나타내며, 사람 개발자와 AI 코딩 에이전트가 공통의 구조적 컨텍스트를 유지하기 위한 정보 소스로 활용합니다.

## 3층 지식 구조
이 프로젝트는 AI와 사람의 효율적인 협업 및 장기적 의사결정 추적을 위해 다음과 같은 3층 문서 체계를 준수합니다.
- **CLAUDE.md**: 에이전트의 구동 세션 시작 시 항상 로드되어야 하는 기본적인 팀 코딩 컨벤션 및 개발 규칙을 기술합니다.
- **claude-context/architecture.md**: 현재 가동 중인 전체 시스템의 컴포넌트 구조, 리소스 토폴로지, 데이터 흐름 등 동적인 현재 상태(Single Page Summary)를 정의합니다.
- **docs/architecture-decisions.md**: 설계 과정에서 발생한 핵심 아키텍처 결정(ADR)과 기술적 트레이드오프 기록을 순차적으로 누적 보관합니다.

---

## 클러스터 토폴로지
GCP 상에 프로비저닝된 GKE 클러스터의 인프라 구성 정보입니다.

| 항목 | 현재 설정값 | 세부 사양 / 비고 |
| :--- | :--- | :--- |
| **클러스터 이름** | `notiflex-cluster` | GKE Standard (Zonal) |
| **리전 / 존** | `asia-northeast3 / asia-northeast3-a` | 서울 리전 |
| **노드풀 구성** | `default-pool` (e2-medium x 2 노드) | Spot VM 활용 (비용 최적화) |
| **활성화된 GKE 기능** | `Gateway API`, `Workload Identity` | `secrets-store-gke.csi.k8s.io` addon |

---

## 컴포넌트 다이어그램
외부 사용자 유입부터 최종 캐시 백엔드 및 시크릿 데이터 볼륨까지의 인플로우 아키텍처 관계도입니다.

```
[외부 트래픽]
       │
       ▼
[GKE L7 외부 로드밸런서 (gke-l7-regional-external-managed)]
       │ (HealthCheckPolicy: /health:8080 감시)
       ▼
[HTTPRoute (notiflex-route)] ───► (가중치 기반 라우팅 제어)
       │
       ├─► [Stable Service (notiflex-api)] ──► Pod A (Stable: v0.5.0) ──┐
       │                                                                ├─► [Valkey Standalone (valkey-primary)]
       └─► [Canary Service (notiflex-api-preview)] ──► Pod B (Canary) ──┘   (ID 생성 동기화: INCR)
                                                  ▲
                                                  │ (볼륨 마운트)
                                         [SecretProviderClass]
                                                  ▲
                                                  │ (Workload Identity)
                                     [Google Secret Manager]
                                    (secret: valkey-password)
```

---

## 배포 파이프라인
Notiflex 서비스의 선언적 GitOps 배포 흐름입니다.

```
[코드 작성] ──► [Git Push (main branch)] ──► [GitHub Actions (CI)]
                                                    │
                                                    ▼ (Docker Build)
[GKE K8s Cluster] ◄── [ArgoCD Auto Sync] ◄── [Artifact Registry]
  (Argo Rollouts)   (notiflex-smb App)
```
- **빌드 파이프라인**: main 브랜치 푸시 시 GitHub Actions에서 태그를 기반으로 이미지를 빌드한 뒤 `asia-northeast3-docker.pkg.dev`에 업로드하고, Argo CD가 감지해 Rollout 매니페스트 변경분을 자동으로 동기화합니다.
- **Canary 롤아웃 전략**: Argo Rollouts가 새 버전을 감지하면 `setWeight: 20` (30s) ➔ `50` (30s) ➔ `80` (30s) ➔ `100` (stable 승격) 단계를 밟으며 무중단 배포를 실행합니다.

---

## 관측 가능성
서비스 안정성 확보 및 디버깅을 위해 클러스터 내에 구축된 관측성 모니터링 스택입니다.

| 도구 이름 | 역할 및 수집 범위 | 리소스 최적화 (Requests) |
| :--- | :--- | :--- |
| **Prometheus** | Pod 메트릭, HTTP 요청 수, CPU/MEM 시계열 수집 | `cpu: 5m`, `memory: 256Mi` (ch6 최적화) |
| **Grafana** | 수집된 메트릭 시각화 대시보드 제공 (포트: 3000) | `cpu: 5m`, `memory: 128Mi` (ch6 최적화) |
| **Loki** | 분산 환경 Pod 내 로그 중앙 집중화 백엔드 | `cpu: 5m`, `memory: 128Mi` (ch6 최적화) |
| **Fluent Bit** | 노드별 컨테이너 로그 스트리밍 (DaemonSet) | `cpu: 5m`, `memory: 64Mi` (ch6 최적화) |

---

## 주요 네임스페이스
클러스터의 네임스페이스 구조 및 가동 중인 워크로드 요약입니다.

| 네임스페이스 | 주요 워크로드 | 리소스 상세 |
| :--- | :--- | :--- |
| **`notiflex`** | Notiflex API Pod, Valkey, GKE Gateway, HTTPRoute | 사용자 서비스 실행 및 분산 상태 공유 |
| **`argocd`** | ArgoCD Application Controller, Repo Server 등 | 선언적 GitOps 릴리스 자동화 |
| **`argo-rollouts`** | Argo Rollouts Controller Daemon | Canary / BlueGreen 배포 오케스트레이션 |
| **`monitoring`** | Prometheus Operator, Grafana, Loki, Fluent Bit | 통합 관측 가능성 스택 가동 |
