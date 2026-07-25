# Notiflex 플랫폼 신규 엔지니어 온보딩 가이드

## 1. 플랫폼 개요
Notiflex는 B2B 알림 SaaS 플랫폼으로, Kubernetes(GKE) 기반의 멀티 테넌시 및 이벤트 기반 아키텍처로 구동됩니다.

- **기초 언어/런타임**: Go `1.25`
- **배포 도구**: GitHub Actions (CI) + Argo CD (GitOps CD)
- **무중단 배포**: Gateway API + Argo Rollouts (Canary / Blue-Green)
- **공유 캐시/상태**: Valkey
- **비동기 메시징**: Strimzi Kafka (KRaft 모드)
- **관측 가능성**: Prometheus(메트릭) + Loki(로그) + Tempo(트레이스) + Grafana

## 2. 저장소 구조
```text
notiflex-platform/
├── app/                      # Go API 서버 소스 및 Dockerfile
├── k8s/                      # Kubernetes 선언적 매니페스트
│   ├── smb/                  # SMB 테넌트 Rollout 및 Service
│   ├── enterprise/           # Enterprise 테넌트 Rollout
│   ├── gateway/              # Gateway API & HTTPRoute
│   ├── monitoring/           # 모니터링 및 알림 설정
│   ├── kafka/                # Strimzi Kafka 노드풀, 클러스터, 토픽
│   └── secret/               # SecretProviderClass
├── argocd/                   # Argo CD Root Application 및 하위 App
│   ├── root-app.yaml         # App of Apps 패턴의 루트 애플리케이션
│   └── apps/                 # 개별 Argo CD Application들
├── command-guardrails/       # 위험 작업 실행 및 검증 절차서
├── docs/                     # 프로젝트 문서 (JOURNEY.md, ADR, 온보딩)
└── .claude/                  # AI 행동 규칙 (settings.local.json 등)
```

## 3. 개발 및 배포 워크플로우
1. 코드 수정을 완료하면 `git push origin main`을 수행합니다.
2. GitHub Actions CI가 자동으로 이미지를 빌드하여 Google Artifact Registry에 푸시합니다.
3. Argo CD `root-app`이 Sync Wave 순서에 맞춰 클러스터 상태를 자동으로 동기화합니다.
   - **Wave 1**: 인프라/관측 가능성 (`monitoring`)
   - **Wave 2**: 플랫폼 및 테넌트 애플리케이션 (`notiflex-smb`, `notiflex-enterprise`, `kafka`)

## 4. 위험 작업 가이드라인
Kafka Topic 삭제, Namespace 삭제, CronJob 수동 실행 등의 작업 시 반드시 `command-guardrails/` 디렉터리의 사전 확인 -> 실행 -> 사후 검증 절차서를 준수합니다.
