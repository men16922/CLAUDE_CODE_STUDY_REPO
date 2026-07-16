# Architecture Decision Records

이 문서는 Notiflex SaaS 플랫폼을 구축하고 운영하면서 내린 중요한 아키텍처 결정사항들을 기록합니다. 각 결정의 맥락과 트레이드오프를 투명하게 보존하여 향후 시스템 개선과 협업의 근거로 사용합니다.

## ADR-001: GitOps 도구로 Argo CD 채택 (3장)
**시점**: 2026-07 / **결정**: Kubernetes 매니페스트 배포 자동화 및 Git 상태 동기화를 위해 Argo CD를 도입합니다.
**이유**:
- Git 저장소를 Single Source of Truth(SSOT)로 관리하여 인프라의 선언적 관리를 보장합니다.
- 변경 사항 발생 시 자동 Sync 정책(`syncPolicy.automated`)을 통해 클러스터 상태와 Git 저장소의 상태를 일치시킵니다.
- 배포 상태를 웹 콘솔을 통해 시각적으로 쉽게 모니터링할 수 있어 트러블슈팅과 운영 효율성을 높입니다.
- Git 기반의 승인 및 롤백 워크플로우를 단순화합니다.

## ADR-002: CI 도구로 GitHub Actions 사용 (3장)
**시점**: 2026-07 / **결정**: 코드 빌드 및 태깅, 매니페스트 변경 감지 자동화를 위해 GitHub Actions CI 파이프라인을 구축합니다.
**이유**:
- GitHub 비공개 저장소와 자연스럽게 통합되어 추가 인프라 구축 비용이 들지 않습니다.
- GCP 서비스 계정 키(`GCP_SA_KEY`)를 통해 Artifact Registry로의 Docker 이미지 푸시 및 매니페스트 태그 업데이트 자동화를 매끄럽게 수행합니다.
- 선언적 YAML 기반 설정으로 CI 워크플로우의 버전을 쉽게 관리합니다.

## ADR-003: 메트릭 모니터링을 위해 Prometheus 및 Grafana 구성 (4장)
**시점**: 2026-07 / **결정**: 클러스터 자원 사용량 모니터링 및 시각화를 위해 kube-prometheus-stack을 도입합니다.
**이유**:
- Prometheus는 시계열 메트릭 데이터 수집에 표준화된 신뢰성을 제공합니다.
- Grafana를 통해 수집된 메트릭을 대시보드로 시각화하여 클러스터 상태를 직관적으로 관찰할 수 있습니다.
- 실습 환경의 리소스(e2-medium) 제약을 고려하여 Prometheus 100m, Grafana 50m로 자원 Requests를 최적화하여 설치했습니다.
- 향후 Loki 로그 쿼리와 연동하여 통합 모니터링 뷰를 제공합니다.

## ADR-004: 로그 수집을 위해 Loki 및 Fluent Bit 도입 (4장)
**시점**: 2026-07 / **결정**: 분산된 Pod 로그를 중앙 집중화하기 위해 Loki와 Fluent Bit를 도입합니다.
**이유**:
- Loki는 텍스트 로그의 인덱싱 비용을 대폭 줄여 리소스 사용량이 적은 효율적인 로그 백엔드를 구축할 수 있게 합니다.
- Fluent Bit는 DaemonSet 형태로 배포되어 리소스를 매우 적게 사용하며, 노드 상의 컨테이너 로그를 Loki로 효율적으로 스트리밍합니다.
- Grafana와 완전하게 호환되어 메트릭과 로그 데이터를 하나의 대시보드에서 분석할 수 있습니다.

## ADR-005: 임계값 기반 알림을 위해 PrometheusRule 구성 (4장)
**시점**: 2026-07 / **결정**: 서비스 장애 전조 현상을 감지하고 Alertmanager로 경보를 전달하기 위해 PrometheusRule을 설정합니다.
**이유**:
- Pod의 비정상적인 재시작(`PodRestartTooMany` 등)을 사전에 정의된 임계값 규칙으로 즉시 탐지할 수 있습니다.
- Kubernetes 네이티브 CRD 설정을 사용하여 알림 룰을 GitOps 파이프라인으로 관리할 수 있습니다.
- Alertmanager 연동을 통해 이메일이나 슬랙 등 외부 채널로의 알림 확장을 가능케 합니다.

## ADR-006: 외부 진입점은 Gateway API로 구성 (5장)
**시점**: 2026-07 / **결정**: GKE의 관리형 Gateway API를 사용하여 Notiflex 외부 접속 경로와 라우팅 규칙을 수립합니다.
**이유**:
- GKE 관리형 GatewayClass(`gke-l7-regional-external-managed`)를 활용하므로 별도의 NGINX Ingress Controller 등 자체 컨트롤러를 설치 및 관리할 필요가 없어 리소스를 절약합니다.
- Gateway와 HTTPRoute 리소스를 통해 인프라 제어 영역과 애플리케이션 라우팅 규칙을 효과적으로 분리합니다.
- HTTPRoute의 weight를 조절하여 Argo Rollouts의 트래픽 라우팅 및 점진적 배포 전환 기능과 매끄럽게 연동할 수 있습니다.
- HealthCheckPolicy를 이용해 로드밸런서 수준에서 `/health:8080` 경로를 직접 감시하도록 제어하여 안정적인 백엔드 라우팅을 확보합니다.

## ADR-007: 무중단 배포는 Argo Rollouts Blue/Green으로 구성 (5장)
**시점**: 2026-07 / **결정**: 기본 Deployment 리소스를 Argo Rollouts의 Rollout CRD로 전면 대체하고 Blue/Green 전략을 도입합니다.
**이유**:
- 배포 프로세스와 트래픽 전환 시점을 안전하게 분리합니다.
- `previewService`를 통해 운영 트래픽과 무관한 환경에서 새 버전의 `/version` 및 `/health`를 미리 완벽히 검증할 수 있습니다.
- 기존 ReplicaSet(Blue)을 전환 후 일정 시간 유지하여, 새 버전에서 에러 발생 시 즉각적으로 기존 ReplicaSet으로 즉시 승격 롤백(`undo`)할 수 있는 안정성을 확보합니다.
- `autoPromotionSeconds: 30` 대기 후 자동으로 stable 승격이 실행되어 효율적인 릴리스 프로세스를 자동화합니다.
- **트레이드오프**: 전환 중 일시적으로 리소스 사용량이 2배 필요하지만 현재 Notiflex API의 규모(2 replicas, e2-medium 노드 2대)에서는 감당할 수 있는 수준입니다.
