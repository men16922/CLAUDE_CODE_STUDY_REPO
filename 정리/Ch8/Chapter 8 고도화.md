# 8장. 고도화

## 8장의 목표

7장에서는 멀티 노드풀로 워크로드를 분리하고 App of Apps로 애플리케이션을 체계적으로 관리하며 Namespace 기반 멀티 테넌시로 고객별 격리 환경을 만들었습니다.

규모는 확장됐지만 시스템의 처리 흐름과 운영 방식에는 다음 한계가 남아 있습니다.

- API가 알림 작업을 직접 끝낼 때까지 기다리므로 요청이 몰리면 응답이 느려지고 타임아웃이 발생합니다.
- API → Valkey → Kafka → Consumer처럼 호출 경로가 길어지면 장애가 발생한 구간을 로그만으로 찾기 어렵습니다.
- 헬스체크나 일일 통계와 같은 반복 작업을 사람이 직접 실행하면 누락과 운영 편차가 생깁니다.
- Kafka Topic 삭제, CronJob 수동 실행, Namespace 삭제와 같은 위험 작업은 단순 차단보다 안전한 실행 절차가 필요합니다.

8장에서는 시스템을 다음 단계로 고도화합니다.

1. **Kafka와 Strimzi**로 요청 수신과 처리를 분리해 이벤트 드리븐 아키텍처를 구성합니다.
2. **OpenTelemetry와 Grafana Tempo**로 전체 요청 경로를 분산 트레이싱합니다.
3. **Kubernetes CronJob**으로 주기적인 헬스체크를 자동화합니다.
4. **`command-guardrails/`**에 위험 작업의 사전 확인, 실행, 사후 검증 절차를 문서화합니다.

---

## 8장 전체 흐름

![그림 8-1. 고도화 전체 구조](./images/ch08/fig-8-01-advanced-overview.png)

```mermaid
flowchart LR
C[클라이언트] --> G[Gateway API]
G --> A[Notiflex API]
A --> V[Valkey INCR]
A --> K[Kafka notifications Topic]
K --> W[Consumer 비동기 처리]

A -. OTLP Trace .-> T[Grafana Tempo]
T --> GF[Grafana]
M[Prometheus 메트릭] --> GF
L[Loki 로그] --> GF

CR[Kubernetes CronJob] --> H[/health 헬스체크]
CG[command-guardrails] --> OP[위험 작업의 안전한 실행]
```

이 장에서는 **요청 처리, 장애 추적, 반복 작업, 위험 작업을 각각 독립적인 운영 계층으로 분리하고 GitOps 안에서 일관되게 관리합니다.**

---

# 8.1 Kafka로 구성하는 이벤트 드리븐 아키텍처

<details>
<summary>**8.1 Kafka로 구성하는 이벤트 드리븐 아키텍처**</summary>

## 문제 상황

![그림 8-2. 동기 처리와 Kafka 비동기 처리 비교](./images/ch08/fig-8-02-kafka-event-flow.png)

현재 Notiflex API는 클라이언트의 요청을 받은 뒤 알림 처리를 직접 완료하고 나서 응답합니다.

```
동기 처리
클라이언트 → API 수신 → 알림 처리 → 응답
                              ↑
                     완료될 때까지 대기
```

이 구조에서는 다음 문제가 발생합니다.

- 알림 전송이 오래 걸리면 API 응답도 함께 느려집니다.
- 처리 도중 Pod가 종료되면 아직 완료하지 못한 알림이 유실될 수 있습니다.
- 트래픽 증가 시 API가 요청 수신과 실제 작업을 모두 담당해 확장 단위가 결합됩니다.

이벤트 기반 구조에서는 요청 수신과 실제 처리를 분리합니다.

```
비동기 처리
클라이언트 → API 수신 → Kafka → Consumer
                 ↓          ↓
              즉시 응답   디스크 저장 후 비동기 처리
```

API는 이벤트를 Kafka에 기록한 뒤 즉시 응답하고 Consumer가 별도로 메시지를 처리합니다. Consumer가 일시적으로 중단되더라도 메시지는 Kafka에 남아 있어 다시 이어서 처리합니다.

## 메시징 도구 비교

| 도구 | 특징 | 장점 | 단점 | Notiflex 적합도 |
| --- | --- | --- | --- | --- |
| Kafka + Strimzi | 분산 이벤트 스트리밍, CRD 기반 관리 | 업계 표준, 영속성, 리플레이, GitOps 호환 | 약 512MiB 이상의 리소스와 학습 곡선 필요 | 높음 |
| RabbitMQ | 전통적인 메시지 브로커 | 가볍고 성숙하며 AMQP/MQTT 지원 | 대규모 스트리밍과 장기 이벤트 보관에는 상대적으로 불리 | 중간 |
| NATS | 경량 메시징 | 매우 가볍고 단순함 | 기본 모델은 at-most-once이며 영속성에는 JetStream 필요 | 중간 |
| Redis Streams | Valkey/Redis 내장 스트림 | 추가 브로커 설치 없이 시작 가능 | Kafka 수준의 파티션, 복제, 이벤트 생태계에는 한계 | 낮음 |

### 메시지 모델 비교

```
Kafka
Producer → Topic(디스크 저장) → Consumer Group
- Consumer 장애 후 재처리와 리플레이 가능

RabbitMQ
Producer → Exchange → Queue → Consumer
- 메시지 소비 중심의 전통적 큐 모델

NATS
Publisher → Subject → Subscriber
- 기본은 빠르고 단순한 전달, JetStream으로 영속성 확장

Redis Streams
XADD → Stream → XREAD/XREADGROUP
- 가볍지만 대규모 이벤트 스트리밍 기능은 제한적
```

## Kafka를 선택한 이유

1. **학습 가치**: 실무에서 이벤트 드리븐 아키텍처를 다룰 때 접할 가능성이 높습니다.
2. **GitOps 호환**: Strimzi의 Kafka CRD를 YAML로 선언해 ArgoCD로 관리합니다.
3. **KRaft 지원**: ZooKeeper 없이 단일 Broker로 구성해 학습 환경의 자원을 줄일 수 있습니다.
4. **메시지 영속성**: Consumer가 중단되어도 메시지를 디스크에 보관하고 Offset부터 다시 읽을 수 있습니다.

## 핵심 개념

- **Producer**: Topic으로 메시지를 발행합니다. Notiflex API가 Producer입니다.
- **Consumer**: Topic의 메시지를 가져와 처리합니다. 현재는 API Pod 내부의 Consumer Goroutine입니다.
- **Topic**: 메시지가 저장되는 논리적 카테고리입니다. 알림 이벤트는 `notifications` Topic에 저장합니다.
- **Partition**: Topic을 병렬 처리하는 단위입니다. Consumer Group 안에서는 한 Partition을 한 Consumer가 담당합니다.
- **Consumer Group**: 같은 그룹의 Consumer들이 Partition을 나눠 처리합니다. Consumer가 추가되거나 제거되면 Kafka가 Partition을 재배치합니다.
- **Offset**: Consumer Group이 Partition에서 어디까지 처리했는지를 나타냅니다. Commit한 Offset 이후부터 처리를 재개합니다.
- **KRaft**: ZooKeeper 없이 Kafka 내부의 Raft 합의로 메타데이터를 관리하는 모드입니다.
- **Strimzi Operator**: Kubernetes에서 Kafka, KafkaNodePool, KafkaTopic 등의 CRD를 관리합니다.

## Kafka 설치 및 구성

Strimzi Operator를 `kafka` Namespace에 설치하고 Kafka Broker는 7장에서 만든 `worker-pool`에 배치합니다.

```bash
kubectl --context gke-sysnet4admin_book_gitaiops create namespace kafka

helm repo add strimzi https://strimzi.io/charts/

helm install strimzi strimzi/strimzi-kafka-operator \
  -n kafka \
  --set 'nodeSelector.cloud\.google\.com/gke-nodepool=worker-pool'

kubectl --context gke-sysnet4admin_book_gitaiops get pods -n kafka
```

```
NAME                                         READY   STATUS    RESTARTS
strimzi-cluster-operator-...                  1/1     Running   0
```

### KRaft 기반 KafkaNodePool

Strimzi 0.51 이상에서는 Kafka 4.x와 `kafka.strimzi.io/v1` API를 사용하며 노드의 역할과 리소스는 `KafkaNodePool`에 선언합니다.

```yaml
# k8s/kafka/kafka-nodepool.yaml
apiVersion: kafka.strimzi.io/v1
kind: KafkaNodePool
metadata:
  name: broker
  namespace: kafka
  labels:
    strimzi.io/cluster: notiflex-kafka
spec:
  replicas: 1
  roles:
    - controller
    - broker
  storage:
    type: jbod
    volumes:
      - id: 0
        type: persistent-claim
        size: 10Gi
        deleteClaim: true
  resources:
    requests:
      cpu: 200m
      memory: 512Mi
    limits:
      cpu: 500m
      memory: 1Gi
  template:
    pod:
      affinity:
        nodeAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
            nodeSelectorTerms:
              - matchExpressions:
                  - key: cloud.google.com/gke-nodepool
                    operator: In
                    values:
                      - worker-pool
```

```yaml
# k8s/kafka/kafka-cluster.yaml
apiVersion: kafka.strimzi.io/v1
kind: Kafka
metadata:
  name: notiflex-kafka
  namespace: kafka
  annotations:
    strimzi.io/kraft: enabled
    strimzi.io/node-pools: enabled
spec:
  kafka:
    version: 4.1.0
    listeners:
      - name: plain
        port: 9092
        type: internal
        tls: false
    config:
      offsets.topic.replication.factor: 1
      transaction.state.log.replication.factor: 1
      transaction.state.log.min.isr: 1
  entityOperator:
    topicOperator:
      resources:
        limits:
          memory: 512Mi
```

```bash
kubectl --context gke-sysnet4admin_book_gitaiops apply \
  -f k8s/kafka/kafka-nodepool.yaml

kubectl --context gke-sysnet4admin_book_gitaiops apply \
  -f k8s/kafka/kafka-cluster.yaml

kubectl --context gke-sysnet4admin_book_gitaiops wait \
  kafka/notiflex-kafka \
  --for=condition=Ready \
  --timeout=300s \
  -n kafka

kubectl --context gke-sysnet4admin_book_gitaiops get kafka -n kafka
```

Kafka 상태에서 ZooKeeper Replica 항목이 비어 있고 `READY=True`이면 KRaft 모드로 동작합니다.

## notifications Topic 생성

Topic은 세 개의 Partition으로 생성합니다.

```yaml
# k8s/kafka/topic-notifications.yaml
apiVersion: kafka.strimzi.io/v1
kind: KafkaTopic
metadata:
  name: notifications
  namespace: kafka
  labels:
    strimzi.io/cluster: notiflex-kafka
spec:
  partitions: 3
  replicas: 1
```

```bash
kubectl --context gke-sysnet4admin_book_gitaiops apply \
  -f k8s/kafka/topic-notifications.yaml

kubectl --context gke-sysnet4admin_book_gitaiops get kafkatopic -n kafka
```

```
NAME            CLUSTER          PARTITIONS   REPLICATION FACTOR   READY
notifications   notiflex-kafka   3            1                    True
```

> 🔗 **GCP 콘솔 직접 확인 (로그인 필요)**:  
> [Google Cloud Console - GKE 워크로드 관리 화면](https://console.cloud.google.com/kubernetes/workload_/gcloud/asia-northeast3-a/notiflex-cluster?project=claude-study-501117)에서 `kafka` 네임스페이스에 설치된 Strimzi Cluster Operator 및 KRaft 기반 Kafka Broker Pod(`worker-pool` 배치)를 직접 시각적으로 확인하실 수 있습니다.


Partition이 세 개이면 같은 Consumer Group에서 최대 세 Consumer가 동시에 유효하게 작업합니다.

```
Consumer 1개: 세 Partition을 한 Consumer가 처리
Consumer 2개: 한 Consumer가 2개, 다른 Consumer가 1개 처리
Consumer 3개: 각 Consumer가 1개씩 처리
Consumer 4개: 한 Consumer는 할당받을 Partition이 없어 대기
```

## Notiflex를 Producer와 Consumer로 변경

Go Kafka Client로 Sarama를 추가합니다.

```bash
cd app/
go get github.com/IBM/sarama
go mod tidy
```

Rollout에는 Kafka Broker의 Cluster DNS를 환경변수로 전달합니다.

```yaml
env:
  - name: KAFKA_BROKER
    value: notiflex-kafka-kafka-bootstrap.kafka.svc.cluster.local:9092
```

`/id` 요청의 처리 흐름을 다음과 같이 변경합니다.

```
클라이언트 → /id 요청
          → Valkey INCR로 ID 생성
          → Kafka notifications Topic에 이벤트 발행
          → 즉시 HTTP 응답
          → Consumer Goroutine이 비동기로 메시지 처리
```

학습 환경에서는 Producer와 Consumer를 같은 Pod에 구현하지만 프로덕션에서는 Consumer를 별도 Deployment로 분리해 독립적으로 Scale-out하는 것이 일반적입니다.

```bash
gcloud builds submit app/ \
  --tag=asia-northeast3-docker.pkg.dev/PROJECT_ID/notiflex/api:v0.6.0

cd notiflex-platform
sed -i 's/v0.5.0/v0.6.0/' k8s/smb/rollout.yaml

git add app/ k8s/
git commit -m "feat: add Kafka event-driven messaging with Strimzi"
git push origin main
```

ArgoCD가 변경을 감지하고 Canary 배포를 진행합니다.

```bash
kubectl --context gke-sysnet4admin_book_gitaiops \
  argo rollouts get rollout notiflex-api -n notiflex -w
```

### 동작 확인

```bash
curl http://$GATEWAY_IP/id
curl http://$GATEWAY_IP/id

kubectl --context gke-sysnet4admin_book_gitaiops \
  logs -l app=notiflex-api -n notiflex --tail=10 \
  | grep "Kafka consumer"
```

```
Kafka consumer: received on notifications: {"id":13,"timestamp":"..."}
Kafka consumer: received on notifications: {"id":14,"timestamp":"..."}
```

이전에는 ID 생성과 알림 처리가 모두 끝나야 응답했지만 이제는 ID를 생성하고 Kafka에 기록한 뒤 바로 응답합니다.

## Consumer가 재시작되는 경우

Kafka는 Consumer Group이 어디까지 읽었는지 Offset으로 기록합니다. Consumer가 메시지 처리를 마치고 Offset을 Commit하면, 재시작 후 마지막 Commit 지점 다음 메시지부터 읽습니다.

```
Partition 0: [msg-1] [msg-2] [msg-3] [msg-4] [msg-5]
                                  ↑
                          committed offset

Consumer 재시작 → msg-4부터 처리
```

메시지는 설정된 보관 기간 동안 디스크에 남아 있으므로 Offset을 되돌려 과거 이벤트를 다시 처리할 수도 있습니다.

## Kafka Connect

Kafka Connect는 애플리케이션 코드를 직접 작성하지 않고 외부 시스템과 Kafka를 연결하는 프레임워크입니다.

- **Debezium Source Connector**: MySQL, PostgreSQL, MongoDB 등의 변경 이벤트를 CDC 방식으로 Kafka에 전송합니다.
- **JDBC Source/Sink Connector**: 데이터베이스 테이블과 Kafka Topic을 동기화합니다.
- **S3/BigQuery Sink Connector**: Kafka 메시지를 Object Storage나 분석 플랫폼에 적재합니다.

현재 Notiflex는 Go 애플리케이션이 Producer와 Consumer 역할을 직접 수행하므로 Connector가 필요하지 않습니다. 데이터베이스 연동이나 외부 분석 시스템 통합이 생길 때 검토합니다.

## 프로덕션 고려사항

- Broker를 세 대 이상으로 구성하고 Replication Factor와 `min.insync.replicas`를 조정합니다.
- Consumer를 별도 Deployment로 분리하고 HPA 또는 KEDA로 확장합니다.
- TLS, SASL 인증, NetworkPolicy를 적용합니다.
- Topic 보관 기간, DLQ, 재처리 정책을 명시합니다.
- 단일 Broker와 `replicas: 1`은 학습 환경 전용 구성입니다.

</details>

---

# 8.2 Tempo로 구성하는 분산 트레이싱

<details>
<summary>**8.2 Tempo로 구성하는 분산 트레이싱**</summary>

## 문제 상황

Kafka를 도입하면서 하나의 요청이 여러 컴포넌트를 거치게 되었습니다.

```
API /id 핸들러
├── Valkey INCR
├── Kafka Produce
└── HTTP 응답

Kafka Topic
└── Consumer 비동기 처리
```

알림 처리가 실패하면 API의 Kafka 발행이 실패했는지, Kafka에는 들어갔지만 Consumer가 처리하지 못했는지, 어느 호출 구간이 느렸는지를 Pod별 로그만으로 판단하기 어렵습니다.

## 관측 가능성의 세 요소

![그림 8-3. 메트릭·로그·트레이스의 연결](./images/ch08/fig-8-03-observability-trace.png)

| 요소 | 도구 | 도입 시점 | 역할 |
| --- | --- | --- | --- |
| 메트릭 | Prometheus + Grafana | 4장 | 무엇이 이상한지 숫자로 확인 |
| 로그 | Loki + Fluent Bit | 4장 | 어떤 오류가 발생했는지 텍스트로 확인 |
| 트레이스 | Grafana Tempo | 8장 | 요청이 어느 경로에서 지연되거나 실패했는지 확인 |

메트릭과 로그에 트레이스를 추가하면 하나의 요청이 시스템 전체를 통과한 경로와 각 구간의 처리 시간을 연결해서 볼 수 있습니다.

## Trace와 Span

- **Trace**: 하나의 요청이 시스템을 통과하는 전체 경로입니다.
- **Span**: Trace를 구성하는 개별 작업 구간입니다. 시작 시점, 종료 시점, 상태, 속성을 기록합니다.
- **TraceID**: 하나의 요청을 식별하는 고유 ID입니다.
- **OpenTelemetry**: 트레이스, 메트릭, 로그 수집을 위한 벤더 중립 표준입니다.
- **OTLP**: OpenTelemetry 데이터를 전송하는 프로토콜입니다. 기본 포트는 gRPC 4317, HTTP 4318입니다.

```
Trace: /id 요청 전체 55ms
├── handle-id       55ms
├── valkey-incr      3ms
├── kafka-produce   12ms
└── http-response    1ms
```

코드에서 외부 호출 전에 Span을 시작하고 호출이 끝나면 Span을 종료합니다. 여러 Span이 하나의 TraceID 아래에 모이면 전체 요청 경로가 됩니다.

## 분산 트레이싱 도구 비교

| 도구 | 특징 | 장점 | 단점 | Notiflex 적합도 |
| --- | --- | --- | --- | --- |
| Grafana Tempo | Grafana 네이티브 트레이싱 | Grafana 통합, 경량, OTLP 지원 | 별도의 자체 UI는 없음 | 높음 |
| Jaeger | CNCF 졸업 프로젝트, 자체 UI | 독립 UI와 성숙한 생태계 | 대규모 저장소 구성 시 Elasticsearch/Cassandra 등이 필요 | 중간 |
| Zipkin | 원조 분산 트레이싱 도구 | 설치가 단순하고 자체 UI 제공 | 기능과 커뮤니티가 상대적으로 축소 | 낮음 |

### Tempo를 선택한 이유

1. 4장에서 설치한 Grafana에 데이터 소스만 추가하면 됩니다.
2. Prometheus, Loki, Tempo를 Grafana 한곳에서 연결합니다.
3. 단일 바이너리 학습 구성은 적은 CPU와 메모리로 실행됩니다.
4. OpenTelemetry와 OTLP를 기본 지원해 애플리케이션을 특정 Backend에 종속시키지 않습니다.

## Tempo 설치

Tempo는 `monitoring` Namespace에 설치하고 7장에서 만든 `ops-pool`에 배치합니다. OTLP gRPC Receiver를 포트 4317에서 활성화합니다.

```bash
helm repo add grafana https://grafana.github.io/helm-charts

helm install tempo grafana/tempo -n monitoring \
  --set tempo.receivers.otlp.protocols.grpc.endpoint="0.0.0.0:4317" \
  --set tempo.resources.requests.cpu=25m \
  --set tempo.resources.requests.memory=128Mi \
  --set 'nodeSelector.cloud\.google\.com/gke-nodepool=ops-pool'

kubectl --context gke-sysnet4admin_book_gitaiops \
  get pods -n monitoring -l app.kubernetes.io/name=tempo
```

```
NAME      READY   STATUS    RESTARTS
tempo-0   1/1     Running   0
```

> `grafana/tempo` Chart에서 Deprecated 경고가 표시될 수 있습니다. 학습 환경에서는 단일 바이너리 구성이 충분하지만 대규모 프로덕션 환경에서는 컴포넌트를 분리하는 `tempo-distributed` 구성을 검토합니다.
>

## Grafana에 Tempo 데이터 소스 추가

```bash
kubectl --context gke-sysnet4admin_book_gitaiops \
  port-forward svc/kube-prometheus-grafana \
  -n monitoring 3000:80 &

curl -X POST http://admin:prom-operator@localhost:3000/api/datasources \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Tempo",
    "type": "tempo",
    "url": "http://tempo:3100",
    "access": "proxy",
    "isDefault": false
  }'

kill %1
```

## OpenTelemetry를 사용하는 이유

애플리케이션이 Tempo SDK에 직접 의존하면 Backend를 Jaeger나 다른 서비스로 변경할 때 코드 전체를 수정해야 합니다.

```
애플리케이션 → OpenTelemetry SDK → OTLP → Tempo
                                         ├→ Jaeger
                                         ├→ Datadog
                                         └→ 기타 OTLP Backend
```

애플리케이션은 OTLP로 전송한다는 사실만 압니다. 실제 수신 Backend는 `OTEL_EXPORTER_OTLP_ENDPOINT` 환경변수로 교체합니다.

## Notiflex에 OpenTelemetry 추가

```bash
cd app/

go get go.opentelemetry.io/otel
go get go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc
go get go.opentelemetry.io/otel/sdk/trace
go mod tidy
```

핵심 변경 사항은 다음과 같습니다.

- OTLP gRPC Exporter로 `TracerProvider`를 초기화합니다.
- `/id` HTTP Handler에 최상위 `handle-id` Span을 추가합니다.
- Valkey 호출에는 `valkey-incr` Span을 추가합니다.
- Kafka 발행에는 `kafka-produce` Span을 추가합니다.
- Tempo 주소는 환경변수로 전달합니다.

```yaml
env:
  - name: OTEL_EXPORTER_OTLP_ENDPOINT
    value: tempo.monitoring.svc.cluster.local:4317
```

```bash
gcloud builds submit app/ \
  --tag=asia-northeast3-docker.pkg.dev/PROJECT_ID/notiflex/api:v0.7.0

cd notiflex-platform
sed -i 's/v0.6.0/v0.7.0/' k8s/smb/rollout.yaml

git add app/ k8s/ helm-values/
git commit -m "feat: add OpenTelemetry tracing with Tempo"
git push origin main

kubectl --context gke-sysnet4admin_book_gitaiops \
  argo rollouts get rollout notiflex-api -n notiflex -w
```

## 트레이스 확인

```bash
curl http://$GATEWAY_IP/id
curl http://$GATEWAY_IP/id
```

Grafana의 **Explore**에서 데이터 소스를 Tempo로 선택하고 `Service Name=notiflex-api`로 조회합니다.

```
notiflex-api /id [55ms]
├── valkey-incr     [3ms]
├── kafka-produce  [12ms]
└── http-response   [1ms]
```

각 Span의 소요 시간을 비교하면 어느 외부 호출에서 지연이 발생했는지 한눈에 드러납니다.

## Span은 어디에 넣어야 하는가

모든 함수에 Span을 만들면 Trace가 지나치게 길어져 오히려 읽기 어렵습니다. Span은 시스템 또는 책임 경계를 넘는 지점에 배치합니다.

- **HTTP Handler**: 요청이 들어오는 전체 진입점과 총 처리 시간을 기록합니다.
- **외부 시스템 호출**: Valkey, Kafka, Database, 외부 API 호출을 기록합니다.
- **비동기 처리**: Consumer가 메시지를 수신하고 처리하는 구간을 기록합니다.
- 단순 JSON 인코딩, 문자열 포맷팅과 같은 내부 로직에는 기본적으로 넣지 않습니다.

## 메트릭·로그·트레이스 연결

관측 가능성의 세 요소가 모두 구성되었습니다.

1. Grafana의 Prometheus Dashboard에서 `/id` Error Rate가 증가했음을 확인합니다.
2. 동일 시간대의 Loki 로그에서 `Kafka produce failed` 같은 오류를 확인합니다.
3. 로그의 TraceID로 Tempo에서 해당 요청을 조회합니다.
4. `kafka-produce` Span이 지연되거나 Timeout으로 실패한 것을 확인합니다.
5. Kafka Broker의 CPU와 Memory 메트릭을 확인해 근본 원인을 좁힙니다.

```
Prometheus: 무엇이 이상한가
       ↓
Loki: 어떤 오류가 발생했는가
       ↓
Tempo: 요청의 어느 구간에서 막혔는가
       ↓
Grafana: 세 신호를 한곳에서 연결
```

</details>

---

# 8.3 CronJob으로 배치 자동화

<details>
<summary>**8.3 CronJob으로 배치 자동화**</summary>

## 문제 상황

API → Kafka → Consumer로 이어지는 시스템이 정상인지 주기적으로 확인해야 합니다. 현재는 사람이 직접 `curl`로 헬스체크하므로 바쁜 날에는 누락될 수 있고 실행 시간과 결과도 일관되게 남지 않습니다.

Kubernetes에는 반복 작업을 정해진 스케줄에 실행하는 `CronJob` 리소스가 있습니다.

## 주기적 작업 도구 비교

| 방식 | 적합한 경우 | Notiflex 헬스체크 적합도 |
| --- | --- | --- |
| Kubernetes CronJob | 단순 주기 작업 | 최적. 추가 설치가 없고 YAML 한 파일로 GitOps 관리 |
| Argo Workflows | 복잡한 DAG와 작업 간 의존성 | 현재 요구에는 과도하며 별도 CRD가 필요 |
| Airflow | 데이터 파이프라인과 ETL | 별도 서버와 운영 환경이 필요해 과도함 |
| 외부 VM Cron | Kubernetes 외부 스케줄링 | GitOps 관리가 어렵고 클러스터 밖 의존성 발생 |

## Kubernetes CronJob을 선택한 이유

1. Kubernetes 기본 리소스이므로 추가 설치가 없습니다.
2. Linux Cron과 같은 표현식을 사용합니다. `*/5 * * * *`는 5분마다 실행한다는 뜻입니다.
3. Manifest를 Git에 저장하고 ArgoCD로 관리합니다.
4. 실패하면 Job의 재시도 정책에 따라 자동으로 다시 실행합니다.
5. 성공 및 실패 Job History 개수를 제한해 오래된 리소스를 자동 정리합니다.

## 핵심 설정

- **`schedule`**: Cron 표현식입니다. Kubernetes CronJob의 시간대 설정을 확인해야 합니다.
- **`concurrencyPolicy: Forbid`**: 이전 Job이 완료되지 않았으면 새 Job을 동시에 시작하지 않습니다.
- **`successfulJobsHistoryLimit`**: 성공한 Job을 몇 개 보관할지 지정합니다.
- **`failedJobsHistoryLimit`**: 실패한 Job을 몇 개 보관할지 지정합니다.
- **`restartPolicy: OnFailure`**: Container가 실패할 경우 재시도합니다.
- **`nodeSelector`**: API 처리 자원과 분리하기 위해 `ops-pool`에 배치합니다.

## 헬스체크 흐름

```
k8s/smb/healthcheck-cronjob.yaml
        ↓
Schedule: */5 * * * *
        ↓
curl /health
  ├── HTTP 200 → 성공
  └── 그 외    → exit 1, 실패 및 재시도
        ↓
ArgoCD App of Apps가 자동 관리
```

## CronJob 생성

```yaml
# k8s/smb/healthcheck-cronjob.yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: notiflex-healthcheck
  namespace: notiflex
spec:
  schedule: "*/5 * * * *"
  concurrencyPolicy: Forbid
  successfulJobsHistoryLimit: 3
  failedJobsHistoryLimit: 3
  jobTemplate:
    spec:
      template:
        spec:
          nodeSelector:
            cloud.google.com/gke-nodepool: ops-pool
          containers:
            - name: healthcheck
              image: curlimages/curl:latest
              command:
                - /bin/sh
                - -c
                - |
                  HTTP_CODE=$(curl -s -o /dev/null -w '%{http_code}' \
                    http://notiflex-api.notiflex.svc.cluster.local/health)

                  if [ "$HTTP_CODE" = "200" ]; then
                    echo "Health check passed (HTTP $HTTP_CODE)"
                    BODY=$(curl -s \
                      http://notiflex-api.notiflex.svc.cluster.local/health)
                    echo "Response: $BODY"
                    exit 0
                  else
                    echo "Health check FAILED (HTTP $HTTP_CODE)"
                    exit 1
                  fi
              resources:
                requests:
                  cpu: 10m
                  memory: 16Mi
                limits:
                  cpu: 50m
                  memory: 32Mi
          restartPolicy: OnFailure
```

```bash
kubectl --context gke-sysnet4admin_book_gitaiops apply \
  -f k8s/smb/healthcheck-cronjob.yaml

kubectl --context gke-sysnet4admin_book_gitaiops \
  get cronjob -n notiflex
```

```
NAME                   SCHEDULE      SUSPEND   ACTIVE
notiflex-healthcheck   */5 * * * *   False     0
```

> 🔗 **GCP 콘솔 직접 확인 (로그인 필요)**:  
> [Google Cloud Console - GKE CronJob 관리 화면](https://console.cloud.google.com/kubernetes/cronjob/asia-northeast3-a/notiflex-cluster/notiflex/notiflex-healthcheck?project=claude-study-501117)에서 5분마다 `ops-pool` 노드에서 자동 실행되는 헬스체크 배치 Job의 내역을 직접 확인하실 수 있습니다.


첫 스케줄을 기다리지 않고 CronJob에서 일회성 Job을 생성해 즉시 검증합니다.

```bash
kubectl --context gke-sysnet4admin_book_gitaiops create job \
  --from=cronjob/notiflex-healthcheck \
  test-health-001 \
  -n notiflex

kubectl --context gke-sysnet4admin_book_gitaiops wait \
  job/test-health-001 \
  --for=condition=complete \
  --timeout=60s \
  -n notiflex

kubectl --context gke-sysnet4admin_book_gitaiops \
  logs job/test-health-001 -n notiflex
```

```
Health check passed (HTTP 200)
Response: {"status":"ok"}
```

5분 후 자동으로 실행된 Job도 확인합니다.

```bash
kubectl --context gke-sysnet4admin_book_gitaiops get jobs -n notiflex
kubectl --context gke-sysnet4admin_book_gitaiops \
  logs job/notiflex-healthcheck-1712577600 -n notiflex
```

```
Health check passed (HTTP 200)
Response: {"status":"ok"}
```

## 실패하는 경우

`curl`이 HTTP 200을 받지 못하면 Script가 `exit 1`로 종료됩니다. `restartPolicy: OnFailure`와 Job Controller의 Backoff 정책에 따라 재시도하고 최종 실패 Job은 `failedJobsHistoryLimit` 범위에서 남습니다.

```bash
kubectl --context gke-sysnet4admin_book_gitaiops get jobs -n notiflex
kubectl --context gke-sysnet4admin_book_gitaiops describe job <실패한-job> -n notiflex
kubectl --context gke-sysnet4admin_book_gitaiops logs job/<실패한-job> -n notiflex
```

프로덕션에서는 CronJob 실패 메트릭을 Prometheus로 수집하고 Alertmanager와 연결해 자동 알림을 구성합니다.

## GitOps 반영

```bash
cd notiflex-platform

git add k8s/smb/healthcheck-cronjob.yaml
git commit -m "feat: add CronJob for periodic health check"
git push origin main
```

`notiflex-smb` ArgoCD Application이 `k8s/smb/` 디렉터리를 감시하므로 CronJob도 자동으로 배포되고 복구됩니다.

</details>

---

# 8.4 command-guardrails로 위험 작업 절차 정리

<details>
<summary>**8.4 command-guardrails로 위험 작업 절차 정리**</summary>

## 시스템이 복잡해지면

8장까지 오면서 운영 리소스가 크게 늘어났습니다. 이제 다음 작업은 명령 하나면 실행되지만 영향 범위가 큽니다.

- Kafka Topic 삭제
- CronJob 수동 실행 또는 중단
- Tenant Namespace 삭제
- Kafka Broker, Tempo, Node Pool 변경
- GitOps 관리 리소스의 수동 변경

7장의 `.claude/settings.local.json`은 `kubectl delete`처럼 위험한 명령을 차단하거나 사용자 승인을 요구했습니다. 그러나 리소스 삭제 자체가 항상 잘못된 것은 아닙니다. 운영 중에는 Topic이나 Tenant를 실제로 정리해야 할 수도 있습니다.

8장에서는 위험 작업을 무조건 막는 대신 **어떤 순서로 무엇을 확인하고 어떻게 실행하며 실행 후 무엇을 검증할지**를 `command-guardrails/`에 문서화합니다.

![그림 8-4. CronJob 자동화와 command-guardrails 실행 절차](./images/ch08/fig-8-04-automation-guardrails.png)

## command-guardrails의 기본 구조

모든 절차서는 다음 세 단계로 구성합니다.

1. **사전 확인**: 영향 범위, 의존성, 잔여 데이터, 백업과 롤백 가능성을 확인합니다.
2. **실행**: 안전한 순서로 트래픽이나 Producer를 중단한 후 변경합니다.
3. **사후 검증**: 리소스 상태, 로그, 메트릭, 애플리케이션 오류 여부를 확인합니다.

```
위험 명령 요청
    ↓
command-guardrails 절차 검색
    ↓
사전 확인 → 사용자 승인 → 실행 → 사후 검증
    ↓
결과와 변경 이력 기록
```

## 예시 1. Kafka Topic 삭제

```markdown
# Kafka Topic 삭제

## 사전 확인
1. Topic에 미처리 메시지가 없는지 확인한다.
2. 모든 Consumer가 처리를 완료했는지 Offset Lag을 확인한다.
3. 해당 Topic을 사용하는 Producer와 Consumer 목록을 파악한다.
4. 보관 또는 재처리가 필요한 이벤트를 백업한다.

## 실행
1. 관련 Producer를 먼저 중지해 메시지 유입을 차단한다.
2. Consumer가 잔여 메시지를 모두 처리할 때까지 기다린다.
3. Git의 KafkaTopic YAML을 삭제한다.
4. ArgoCD가 KafkaTopic 리소스를 정리하도록 Sync한다.

## 사후 검증
1. Topic이 삭제되었는지 확인한다.
2. 관련 Producer와 Consumer 로그에 오류가 없는지 확인한다.
3. Prometheus와 Grafana에서 Kafka 오류 및 Lag을 확인한다.
```

GitOps로 관리되는 리소스는 직접 `kubectl delete`하는 대신 **Git에서 Manifest 삭제 → PR Review → ArgoCD Prune** 흐름을 우선합니다.

## 예시 2. CronJob 수동 실행

```markdown
# CronJob 수동 실행

## 사전 확인
1. 동일한 CronJob에서 실행 중인 Job이 없는지 확인한다.
2. 중복 실행 시 데이터가 중복 처리되지 않는지 확인한다.
3. 실행 대상 환경과 Namespace를 확인한다.

## 실행
1. CronJob에서 고유한 이름의 일회성 Job을 생성한다.
2. Job의 완료 상태를 기다린다.
3. Job 로그를 확인한다.

## 사후 검증
1. Job의 종료 코드와 처리 결과를 확인한다.
2. 중복 리소스 또는 부작용이 없는지 확인한다.
3. 필요하면 테스트 Job을 정리한다.
```

## 예시 3. Tenant Namespace 삭제

```markdown
# Tenant Namespace 삭제

## 사전 확인
1. 고객 계약 종료와 삭제 승인을 확인한다.
2. 데이터, Secret, PVC, DNS, Gateway Route 목록을 확인한다.
3. 백업과 보존 기간을 확인한다.
4. App of Apps에서 해당 Tenant Application의 위치를 확인한다.

## 실행
1. 외부 트래픽을 차단하고 고객 워크로드를 중단한다.
2. Git에서 Tenant Application과 Manifest를 제거한다.
3. ArgoCD Prune 결과를 확인한 뒤 Namespace를 정리한다.

## 사후 검증
1. Namespace와 종속 리소스가 제거되었는지 확인한다.
2. 공용 Valkey, Kafka, Gateway에 고아 리소스가 없는지 확인한다.
3. 감사 로그와 변경 이력을 기록한다.
```

## settings.local.json과의 차이

| 수단 | 역할 | 예시 |
| --- | --- | --- |
| `settings.local.json` | 명령 실행을 기술적으로 차단하거나 사용자 승인을 요구 | `kubectl delete` 실행 거부, Node Pool 삭제 전 Ask |
| `command-guardrails/` | 실행이 필요한 경우 따라야 할 운영 절차 제공 | 사전 확인 → 승인 → 실행 → 사후 검증 |

7장에서 체험한 방식이 **“하지 마”라는 기술적 통제**라면, 8장의 절차서는 **“해야 한다면 안전하게 이렇게 실행하라”는 운영 통제**입니다.

자연어 규칙만으로 강제하기 어렵거나 실수의 피해가 큰 명령은 권한 통제와 승인 Gate를 유지하고 실제 운영 작업은 절차서에 따라 수행합니다.

## 변경 사항 기록

`/update-docs`로 Kafka, Tempo, CronJob 구성과 세 개의 Architecture Decision을 문서에 누적합니다.

```bash
git status
# M JOURNEY.md
# M docs/architecture-decisions.md
# M claude-context/architecture.md
# A command-guardrails/

git add JOURNEY.md docs/ claude-context/ command-guardrails/
git commit -m "ch8: Kafka, Tempo, CronJob 도입과 위험 작업 절차서 정리"
git push origin main
```

</details>

---

# 8.5 8장 가드레일 살펴보기

<details>
<summary>**8.5 8장 가드레일 살펴보기**</summary>

8장에서는 8.1과 8.2에서 **3-Prompt 패턴인 탐색 → 비교 → 실행**을 사용합니다. 8.3은 Kubernetes 기본 리소스를 다루므로 **탐색 → 실행** 형태의 라이트 가이드를 사용합니다.

| 서브챕터 | 유형 | 참조 파일 | 역할 |
| --- | --- | --- | --- |
| 8.1 탐색/비교 | 탐색과 비교 | `decision-guides/ch8/8.1-messaging.md` | Kafka 추천과 RabbitMQ, NATS, Redis Streams 비교 |
| 8.1 실행 | 실행 | `prompt-guardrails/ch8/8.1-kafka.md` | Strimzi 설치, KRaft Broker, Topic, Producer/Consumer 구성 |
| 8.2 탐색/비교 | 탐색과 비교 | `decision-guides/ch8/8.2-tracing.md` | Tempo 추천과 Jaeger, Zipkin 비교 |
| 8.2 실행 | 실행 | `prompt-guardrails/ch8/8.2-tempo.md` | Tempo 설치와 OpenTelemetry SDK 구성 |
| 8.3 탐색 | 라이트 가이드 | `decision-guides/ch8/8.3-cronjob.md` | Kubernetes CronJob 소개와 운영 기준 |
| 8.3 실행 | 실행 | `prompt-guardrails/ch8/8.3-cronjob.md` | CronJob 생성, 수동 실행, 자동 실행 검증 |
| 8장 마무리 | 실행 절차 | `prompt-guardrails/ch8/command-guardrails-example.md` | Kafka Topic 삭제, CronJob 수동 실행, Tenant Namespace 삭제 절차서 생성 |

## 가드레일에서 주목할 부분

### Strimzi API 버전

Strimzi 0.51 이상에서는 Kafka 4.x와 `kafka.strimzi.io/v1` API를 사용합니다. `KafkaNodePool`에 Broker와 Controller 역할, Storage, Resource, Scheduling 조건을 선언합니다. 버전 호환 표를 확인하지 않으면 오래된 `v1beta2` 예제나 ZooKeeper 기반 설정이 섞일 수 있습니다.

### OpenTelemetry 초기화

Go 애플리케이션에는 다음 항목이 함께 필요합니다.

- OpenTelemetry API와 SDK Package
- OTLP Trace Exporter
- `TracerProvider` 초기화와 종료 시 Flush
- HTTP Handler와 외부 호출 구간의 Span
- `OTEL_EXPORTER_OTLP_ENDPOINT` 환경변수

Package만 추가하고 `TracerProvider`를 등록하지 않으면 Span이 생성되어도 Tempo로 전송되지 않습니다.

### CronJob 중복 실행

헬스체크는 단순하지만 실제 배치 작업은 중복 실행 시 데이터 중복이나 충돌을 일으킬 수 있습니다. `concurrencyPolicy`, Idempotency, Job History, Timeout, Retry 정책을 함께 설계해야 합니다.

### 위험 작업의 절차화

기술적 Deny는 실수를 막아 주지만 운영에 필요한 작업까지 영구적으로 막을 수는 없습니다. 권한 통제, 승인, 사전 확인, 실행, 사후 검증을 결합해야 합니다.

</details>

---

# 8장 완성 아키텍처

```mermaid
flowchart TB
Internet[Internet] --> GW[Gateway API + HTTPRoute]
GW --> API[Notiflex API v0.7.0 · Canary]

API --> VK[Valkey 공유 캐시]
API --> KT[Kafka notifications · 3 Partitions]
KT --> CON[Consumer 비동기 처리]

API -. OTLP .-> TP[Tempo]
PM[Prometheus] --> GF[Grafana]
LK[Loki] --> GF
TP --> GF

CJ[CronJob · 5분] --> API

subgraph GKE[notiflex-cluster]
    DP[default-pool\n시스템·관측 가능성]
    AP[api-pool\nAPI·Valkey]
    WP[worker-pool\nStrimzi·Kafka KRaft]
    OP[ops-pool\nTempo·CronJob]
end
```

| 영역 | 완성된 구조 |
| --- | --- |
| 메시징 | API → Kafka → Consumer 비동기 처리 |
| 관측 가능성 | Prometheus(메트릭) + Loki(로그) + Tempo(트레이스) → Grafana |
| 배포 | ArgoCD App of Apps + Sync Wave + Argo Rollouts Canary |
| 자동화 | Kubernetes CronJob으로 5분 주기 헬스체크 |
| 운영 안전 | `settings.local.json` 기술 통제 + `command-guardrails/` 실행 절차 |

## 8장 핵심 정리

- **Kafka**: API의 동기 알림 처리를 이벤트 기반 비동기 처리로 전환했습니다. API는 Kafka에 메시지를 기록하고 즉시 응답합니다.
- **Tempo**: OpenTelemetry로 요청의 전체 경로와 각 외부 호출 구간의 시간을 추적합니다.
- **CronJob**: 사람이 수행하던 주기적 헬스체크를 Kubernetes 기본 리소스로 자동화했습니다.
- **command-guardrails**: 위험 작업을 무조건 차단하는 것을 넘어 안전하게 실행할 운영 절차를 Git에 누적합니다.

2장에서 시작한 Notiflex는 환경 구성, 배포 파이프라인, 관측 가능성, 무중단 배포, 엔터프라이즈 기반, 규모 확장과 고도화까지 발전했습니다. 다음 장에서는 이 과정을 돌아보고 Git, AI, Ops가 어떻게 연결되었는지 정리합니다.
