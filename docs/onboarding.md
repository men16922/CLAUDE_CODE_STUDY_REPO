# Notiflex 플랫폼 신규 엔지니어 온보딩 가이드

## 1. 프로젝트 개요
Notiflex는 기업 고객에게 알림 API를 제공하는 B2B SaaS 플랫폼입니다. 테넌트가 API를 호출하면 알림 요청을 생성하고 이벤트를 발행합니다.

- **애플리케이션**: Go 표준 라이브러리 기반 단일 바이너리
- **실행 환경**: GKE (Google Kubernetes Engine)
- **배포 방식**: GitHub Actions (CI) + Argo CD (GitOps CD)
- **외부 진입점**: GKE Gateway API + HTTPRoute
- **배포 전략**: Argo Rollouts Canary (20% → 50% → 80% → 100%)
- **공유 상태**: Valkey standalone 캐시
- **비동기 메시징**: Strimzi Kafka (KRaft 모드)
- **관측 가능성**: Prometheus(메트릭) + Loki(로그) + Tempo(트레이스) + Grafana

### 운영 테넌트
- **SMB**: `notiflex` Namespace의 기본 테넌트. Canary 배포, Valkey 캐시, Kafka 이벤트 발행, CronJob 헬스체크 사용.
- **Enterprise**: `enterprise` Namespace의 고객 격리 테넌트. 별도 Rollout과 Secret 사용.

## 2. 클러스터 및 노드풀 구조
```bash
kubectl --context gke-sysnet4admin_book_gitaiops get nodes \
  -o custom-columns=NAME:.metadata.name,POOL:.metadata.labels.cloud\.google\.com/gke-nodepool
```

| 노드풀 | 용도 | 머신 유형 | 주요 워크로드 |
| --- | --- | --- | --- |
| `default-pool` | 플랫폼 인프라 | `e2-medium` Spot × 2 | Argo CD, Prometheus, Grafana, Loki |
| `api-pool` | API 서빙 | `e2-medium` Spot × 1 | SMB·Enterprise Notiflex API, Valkey |
| `worker-pool` | 데이터 처리 | `e2-standard-2` Spot × 1 | Strimzi Operator, Kafka Broker |
| `ops-pool` | 운영 도구 | `e2-small` Spot × 1 | Tempo, Health Check CronJob |

## 3. Namespace별 워크로드
| Namespace | 주요 Pod | 역할 |
| --- | --- | --- |
| `notiflex` | `notiflex-api`, `valkey-primary`, `notiflex-healthcheck` | SMB API, 공유 캐시, 헬스체크 CronJob |
| `enterprise` | `notiflex-api` | Enterprise 테넌트 API |
| `kafka` | `notiflex-kafka-broker`, `strimzi-cluster-operator` | Strimzi 기반 Kafka KRaft 브로커 |
| `monitoring` | Prometheus, Grafana, Loki, Fluent Bit, Tempo | 메트릭, 로그, 트레이스 통합 관리 |
| `argocd` | Server, Application Controller, Repo Server | GitOps 중앙 배포 통제 |
| `argo-rollouts` | Argo Rollouts Controller | 무중단 Canary & Blue/Green 배포 통제 |

## 4. 저장소 구조
```text
notiflex-platform/
├── CLAUDE.md                 ← 프로젝트 규칙 및 AI 가이드
├── JOURNEY.md                ← 실습 여정 및 의사결정 기록
├── app/                      ← Go 애플리케이션
├── k8s/                      ← Kubernetes 선언적 매니페스트 (smb, enterprise, gateway, kafka 등)
├── argocd/                   ← Argo CD root-app 및 하위 Application
├── command-guardrails/       ← 위험 작업 사전확인/실행/사후검증 절차서
├── docs/                     ← ADR, 온보딩 문서
└── .claude/                  ← 권한 제어 (settings.local.json 등)
```

## 5. 접근 방법

### Argo CD UI
```bash
kubectl --context gke-sysnet4admin_book_gitaiops port-forward svc/argocd-server -n argocd 8080:443
```
- **접속 주소**: `https://localhost:8080`
- **초기 비밀번호 조회**:
  ```bash
  kubectl --context gke-sysnet4admin_book_gitaiops get secret argocd-initial-admin-secret -n argocd -o jsonpath='{.data.password}' | base64 -d
  ```

### Grafana
```bash
kubectl --context gke-sysnet4admin_book_gitaiops port-forward svc/kube-prometheus-grafana -n monitoring 3000:80
```
- **접속 주소**: `http://localhost:3000` (ID: `admin`, PW: `prom-operator`)

## 6. 자주 묻는 Q&A

### Q1. Canary 배포에서 에러율이 증가하면 어떻게 하나요?
즉시 Rollout을 중단하여 Stable 버전으로 트래픽을 복원합니다.
```bash
kubectl --context gke-sysnet4admin_book_gitaiops argo rollouts abort notiflex-api -n notiflex
```

### Q2. 위험 작업(Kafka Topic 삭제, Namespace 삭제 등)을 수행하려면?
`command-guardrails/` 디렉터리의 개별 절차서를 확인하여 사전 확인 → 승인 → 실행 → 사후 검증 3단계를 이행합니다.
