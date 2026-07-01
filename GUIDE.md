# 2장. 개발 환경 구성과 첫 배포

## 2장 목표

이번 장에서는 Notiflex 실습을 시작하기 위한 기본 개발 환경을 구성하고, 첫 번째 API 서버를 GKE에 배포합니다.

### 핵심 목표

- GCP 계정과 무료 크레딧 준비
- 클로드 코드 설치
- gcloud CLI 설치 및 인증
- 프로젝트, 리전, 존 기본값 설정
- Artifact Registry 인증 설정
- GitHub 저장소 구성
- GKE 클러스터 생성
- Notiflex API 서버 빌드와 배포
- 첫 커밋과 문서화 흐름 정리

이 장을 마치면 GKE 위에 Notiflex API 서버를 올리고, 이후 GitOps 실습으로 확장할 수 있는 기반이 준비됩니다.

---

# 2.1 GCP 계정 생성과 무료 크레딧 활용

이 책의 실습은 GCP를 기준으로 진행합니다.

GCP는 신규 계정에 무료 크레딧을 제공하므로, 실습 대부분은 이 범위 안에서 진행할 수 있습니다.

## 사용하는 주요 서비스

| 서비스 | 용도 |
| --- | --- |
| GKE | Kubernetes 클러스터 실행 |
| Artifact Registry | 컨테이너 이미지 저장 |
| Secret Manager | 시크릿 관리 |
| Cloud Build | 이미지 빌드 |
| GitHub | 코드 저장소와 Actions |
| Claude Code | 터미널 기반 AI 작업 도구 |

## 비용 전략

무료 크레딧 안에서 실습을 끝내려면 다음 원칙을 지켜야 합니다.

- 사용하지 않는 클러스터는 삭제한다.
- 리전은 `asia-northeast3`를 사용한다.
- 가능하면 Spot VM을 활용한다.
- 실습 후 리소스를 정리한다.
- GKE 클러스터는 필요할 때 만들고, 끝나면 삭제한다.

> Spot VM은 일반 VM보다 저렴하지만, GCP가 필요하면 회수할 수 있습니다. 이 책의 실습은 Spot VM이 중단되더라도 다시 복구할 수 있도록 설계되어 있습니다.
> 

---

# 2.2 클로드 코드 설치

## 클로드 코드란

클로드 코드는 터미널에서 동작하는 에이전트형 AI입니다.

단순한 코드 자동 완성 도구가 아니라, 명령 실행, 파일 생성, 코드 작성, Git 작업, 오류 분석, 인프라 명령 실행까지 수행할 수 있습니다.

이 책에서는 클로드 코드가 실습의 동료 역할을 합니다.

## 요금제

클로드 코드를 사용하려면 유료 요금제가 필요합니다.

| 요금제 | 실습 가능 여부 | 특징 |
| --- | --- | --- |
| Free | 불가 | 클로드 코드 사용 불가 |
| Pro | 가능 | 장별로 나누어 진행 권장 |
| Max | 권장 | 여러 장을 연속으로 진행하기 좋음 |

Pro 요금제도 실습은 가능하지만, 사용량 제한이 있으므로 장별로 나누어 진행하는 것이 좋습니다.

## 설치와 실행

아래 명령으로 클로드 코드를 설치하고, 실습 가이드 저장소를 클론한 뒤 실행합니다.

```bash
# macOS / Linux에서 클로드 코드 설치
curl -fsSL https://claude.ai/install.sh | bash

# 가이드 저장소 클론
git clone https://github.com/sysnet4admin/_Book_GitAIOps
cd _Book_GitAIOps

# 클로드 코드 실행
# 실습에서는 자동 승인 모드를 사용한다.
claude --dangerously-skip-permissions
```

Windows PowerShell에서는 다음 명령으로 설치합니다.

```powershell
# Windows PowerShell에서 클로드 코드 설치
irm https://claude.ai/install.ps1 | iex
```

> 자동 승인 모드는 파일 생성, 명령 실행, Git 작업을 매번 승인하지 않고 진행할 수 있게 해 줍니다. 단, 실습 전용 GKE 클러스터와 전용 저장소에서만 사용하는 것이 좋습니다.
> 

## 첫 입력

클로드 코드가 정상적으로 실행되면 다음과 같이 입력합니다.

```
안녕! 나는 Notiflex라는 B2B 알림 SaaS 플랫폼의 DevOps 엔지니어야.
이 책을 따라가면서 쿠버네티스 운영 환경을 처음부터 구축하려고 해.
```

그러면 클로드 코드는 현재 저장소의 가드레일을 읽고 실습 흐름을 안내합니다.

## statusline 설정

statusline은 클로드 코드 터미널 하단에 현재 상태를 보여 주는 정보 바입니다.

실습 중 현재 모델, 컨텍스트 사용량, Kubernetes 컨텍스트, 현재 경로 등을 한눈에 확인할 수 있어 유용합니다.

```
/h → Haiku: 빠르고 저렴한 간단 작업
/s → Sonnet: 일반 작업
/o → Opus: 고성능 복잡 작업
```

---

# 2.3 gcloud CLI 설치와 인증

## gcloud CLI가 필요한 이유

GCP 콘솔은 웹 브라우저에서 클릭으로 조작하는 도구입니다.

반면 gcloud CLI는 터미널에서 명령으로 GCP를 제어하는 도구입니다.

이 책에서는 클로드 코드와 함께 인프라 작업을 자동화하기 위해 CLI를 사용합니다.

| 이유 | 설명 |
| --- | --- |
| 자동화 | 명령은 스크립트로 반복할 수 있다. |
| 클로드 코드 연동 | 클로드 코드가 터미널에서 직접 실행할 수 있다. |
| 재현 가능성 | 같은 명령을 실행하면 같은 결과를 얻을 수 있다. |

## 설치 확인, 인증, 기본값 설정

아래 명령으로 gcloud CLI 설치 여부를 확인하고, GCP 인증과 기본 프로젝트 설정을 진행합니다.

```bash
# gcloud CLI 설치 여부 확인
gcloud version

# kubectl이 설치되어 있지 않다면 추가 설치
gcloud components install kubectl

# gcloud CLI 인증
gcloud auth login

# 애플리케이션 기본 인증
gcloud auth application-default login

# 실습 프로젝트 설정
gcloud config set project my-gitaiops-project

# 기본 존 설정
gcloud config set compute/zone asia-northeast3-a

# 기본 리전 설정
gcloud config set compute/region asia-northeast3

# 현재 설정 확인
gcloud config list
```

출력 예시는 다음과 같습니다.

```
[compute]
region = asia-northeast3
zone = asia-northeast3-a

[core]
account = user@example.com
project = my-gitaiops-project
```

`asia-northeast3`는 GCP의 서울 리전입니다.

- 리전: 지역
- 존: 리전 안의 세부 가용 영역
- `asia-northeast3-a`: 서울 리전 안의 가용 영역 중 하나

## Artifact Registry 인증 설정

Artifact Registry는 GCP에서 컨테이너 이미지를 저장하는 전용 레지스트리입니다.

Docker Hub와 비슷하지만, GCP 프로젝트와 IAM 권한에 통합되어 GKE에서 이미지를 가져올 때 더 자연스럽게 연동됩니다.

```bash
# Artifact Registry에 이미지를 푸시할 수 있도록 Docker 인증 설정
gcloud auth configure-docker asia-northeast3-docker.pkg.dev
```

| 항목 | 설명 |
| --- | --- |
| Docker Hub | 공용 컨테이너 레지스트리 |
| Artifact Registry | GCP 전용 컨테이너 레지스트리 |
| 장점 | GCP IAM 기반 접근 제어, GKE와 빠른 연동 |
| 필요 이유 | 빌드한 Notiflex 이미지를 GCP에 저장하기 위해 필요 |

---

# 2.4 클로드 코드와 gcloud의 역할 차이

GitHub Copilot이 코드 작성 보조 도구라면, 클로드 코드는 터미널에서 실제 작업을 수행하는 에이전트에 가깝습니다.

gcloud CLI가 GCP를 제어하는 명령 도구라면, 클로드 코드는 그 명령을 이해하고 실행 흐름을 도와주는 AI 동료입니다.

| 비교 항목 | GitHub Copilot | 클로드 코드 |
| --- | --- | --- |
| 동작 환경 | 에디터 | 터미널 |
| 주요 역할 | 코드 자동 완성, 인라인 제안 | 명령 실행, 파일 생성·수정, Git 작업 |
| 인프라 작업 | 제한적 | 직접 실행 가능 |
| 워크플로 | 코드 작성 보조 | 전체 작업 자동화 |

---

# 2.5 GitHub 저장소 구성

## 목적

Notiflex 프로젝트를 위한 GitHub 저장소를 만들고, 클로드 코드가 프로젝트 구조를 이해할 수 있도록 기본 파일을 생성합니다.

```
notiflex-platform/
├── .gitignore
├── CLAUDE.md
├── README.md
├── app/
├── k8s/
│   └── smb/
└── .github/
    └── workflows/
```

## 주요 파일 역할

| 파일/디렉터리 | 역할 |
| --- | --- |
| `.gitignore` | 빌드 산출물, 인증 파일, OS/IDE 파일 제외 |
| `CLAUDE.md` | 클로드 코드가 읽는 프로젝트 컨텍스트 |
| `README.md` | 프로젝트 설명 |
| `app/` | Go API 서버 코드 |
| `k8s/smb/` | Kubernetes 매니페스트 |
| `.github/workflows/` | 이후 GitHub Actions 파이프라인 위치 |

## 절대 커밋하면 안 되는 것

다음 항목은 공개 저장소에 올리면 안 됩니다.

- 서비스 계정 키
- API 토큰
- OAuth Secret
- DB 비밀번호
- `.env`
- `service-account-*.json`

민감 정보는 GitHub Secrets, 환경 변수, Secret Manager로 관리해야 합니다.

---

# 2.6 GKE 클러스터 생성

## 목적

Notiflex 앱을 올릴 Kubernetes 클러스터를 생성합니다.

| 항목 | 값 |
| --- | --- |
| 클러스터 이름 | `notiflex-cluster` |
| 타입 | GKE Standard |
| 존 | `asia-northeast3-a` |
| 노드 | `e2-medium` 2대 |
| 비용 절감 | Spot VM 사용 |
| Gateway API | 활성화 |
| 디스크 | 30GB |

## 클러스터 생성과 확인

아래 명령으로 GKE 클러스터를 생성하고, `kubectl`이 해당 클러스터를 바라보도록 설정합니다.

```bash
# GKE Standard 클러스터 생성
gcloud container clusters create notiflex-cluster \
  --zone=asia-northeast3-a \
  --machine-type=e2-medium \
  --num-nodes=2 \
  --spot \
  --gateway-api=standard \
  --disk-size=30

# kubectl이 생성한 GKE 클러스터를 바라보도록 설정
gcloud container clusters get-credentials notiflex-cluster \
  --zone=asia-northeast3-a

# kubeconfig 컨텍스트 이름이 너무 길면 짧게 변경
kubectl config rename-context \
  gke_my-gitaiops-project_asia-northeast3-a_notiflex-cluster \
  gke-sysnet4admin_book_gitaiops

# 노드 상태 확인
kubectl get nodes

# Spot VM 적용 여부 확인
kubectl get nodes -o custom-columns=NAME:.metadata.name,SPOT:.metadata.labels.cloud\\.google\\.com/gke-spot

# Gateway API 활성화 여부 확인
kubectl get gatewayclass
```

확인할 항목은 다음과 같습니다.

- 노드 2개가 `Ready` 상태인지 확인
- Spot VM 적용 여부 확인
- GatewayClass 생성 여부 확인

## Spot VM이란

Spot VM은 일반 VM보다 저렴한 대신, GCP가 필요하면 회수할 수 있는 VM입니다.

| 항목 | 설명 |
| --- | --- |
| 장점 | 비용 절감 |
| 단점 | 언제든 중단될 수 있음 |
| 적합한 곳 | 학습, 실습, 비프로덕션 환경 |
| 적합하지 않은 곳 | 안정성이 중요한 프로덕션 환경 |

## kubeconfig 핵심 개념

| 개념 | 설명 |
| --- | --- |
| kubeconfig | kubectl 접속 설정 파일 |
| context | 클러스터, 사용자, 네임스페이스 조합 |
| current-context | 현재 kubectl이 바라보는 대상 클러스터 |

---

# 2.7 Notiflex 앱 빌드와 배포

## 목적

이제 실제 Notiflex API 서버를 만들고 GKE에 배포합니다.

진행 순서는 다음과 같습니다.

1. Go API 서버 작성
2. Dockerfile 작성
3. Artifact Registry에 이미지 빌드 및 푸시
4. Kubernetes 매니페스트 생성
5. 클러스터에 배포
6. 동작 확인

## Go API 서버

Notiflex 앱은 처음에는 단순한 API 서버로 시작합니다.

| 엔드포인트 | 역할 |
| --- | --- |
| `GET /health` | 상태 확인 |
| `GET /id` | 순번 ID와 Pod 이름 반환 |

`/health`는 Kubernetes의 readiness probe와 liveness probe에 사용됩니다.

`/id`는 여러 Pod가 있을 때 어떤 Pod가 응답했는지 확인하는 데 사용됩니다.

## Dockerfile

Go 서버를 컨테이너 이미지로 만들기 위해 Dockerfile을 작성합니다.

이 책에서는 멀티 스테이지 빌드를 사용합니다.

| 단계 | 역할 |
| --- | --- |
| build stage | Go 컴파일러가 있는 이미지에서 바이너리 빌드 |
| runtime stage | `scratch` 이미지에 최종 바이너리만 복사 |

`scratch`는 거의 빈 이미지입니다. OS 패키지와 셸이 없어 이미지가 작고 공격 표면도 줄일 수 있습니다.

## 빌드, 배포, 동작 확인

아래 명령으로 Artifact Registry 저장소를 만들고, 이미지를 빌드한 뒤 GKE에 배포합니다.

```bash
# Artifact Registry 저장소 생성
gcloud artifacts repositories create notiflex \
  --repository-format=docker \
  --location=asia-northeast3 \
  --description="Notiflex container images"

# Cloud Build로 컨테이너 이미지 빌드 및 푸시
gcloud builds submit app/ \
  --tag=asia-northeast3-docker.pkg.dev/my-gitaiops-project/notiflex/api:v0.1.0

# Namespace 먼저 적용
kubectl apply -f k8s/smb/namespace.yaml

# 나머지 Kubernetes 매니페스트 적용
kubectl apply -f k8s/smb/

# Pod 상태 확인
kubectl get pods -n notiflex

# 로컬에서 임시 접근하기 위한 port-forward 실행
kubectl port-forward svc/notiflex-api -n notiflex 8080:80

# 다른 터미널에서 health API 확인
curl -s http://localhost:8080/health

# 다른 터미널에서 id API 확인
curl -s http://localhost:8080/id
```

기대 결과는 다음과 같습니다.

```json
{ "status": "ok" }
```

```json
{
  "id": "1",
  "generated_by": "notiflex-api-xxxxx"
}
```

`generated_by`는 응답한 Pod 이름입니다. 여러 번 호출하면 다른 Pod가 응답할 수도 있습니다.

## Kubernetes 매니페스트 구조

클러스터에 배포할 기본 리소스는 세 가지입니다.

| 리소스 | 역할 |
| --- | --- |
| Namespace | 논리적 격리 공간 |
| Deployment | Pod 복제본 관리 |
| Service | Pod에 접근하는 고정 주소 제공 |

예시 구조는 다음과 같습니다.

```
k8s/smb/
├── namespace.yaml
├── deployment.yaml
└── service.yaml
```

Service 타입은 `ClusterIP`입니다.

즉, 클러스터 내부에서만 접근할 수 있습니다. 외부 접근은 이후 Gateway API나 Ingress를 통해 구성합니다.

---

# 2.8 GitHub에 첫 커밋

## 목적

지금까지 만든 코드와 매니페스트를 GitHub에 올려 2장을 마무리합니다.

커밋 전에는 `JOURNEY.md`를 생성해 실습 진행 상황을 기록합니다.

## JOURNEY.md 역할

| 파일 | 역할 |
| --- | --- |
| `CLAUDE.md` | 프로젝트 컨텍스트와 행동 규칙 |
| `JOURNEY.md` | 실습 진행 상황과 선택 기록 |

`CLAUDE.md`가 AI에게 주는 프로젝트 명함이라면, `JOURNEY.md`는 실습 진행 기록입니다.

## 커밋 대상

```
CLAUDE.md
JOURNEY.md
app/main.go
app/go.mod
app/Dockerfile
k8s/smb/namespace.yaml
k8s/smb/deployment.yaml
k8s/smb/service.yaml
```

커밋과 푸시는 다음과 같이 진행합니다.

```bash
# 전체 변경 사항 스테이징
git add .

# 첫 커밋 생성
git commit -m "Initial commit: Notiflex app, K8s manifests, CLAUDE.md, JOURNEY.md"

# GitHub 원격 저장소에 푸시
git push origin main
```

---

# 2.9 `/update-docs` 스킬 만들기

## 목적

각 장이 끝날 때마다 작업 내용을 문서로 정리하는 커스텀 명령을 만듭니다.

클로드 코드에서는 자주 반복되는 작업을 `/명령어` 형태로 등록할 수 있습니다.

```
/update-docs
```

이 명령은 이번 장에서 변경된 내용을 파악하고, `JOURNEY.md`와 관련 문서를 갱신한 뒤 변경 사항을 커밋하는 역할을 합니다.

## 커스텀 명령 위치

```
.claude/commands/update-docs.md
```

고급 기능이 필요하면 다음 구조도 사용할 수 있습니다.

```
.claude/skills/update-docs/SKILL.md
```

| 방식 | 특징 |
| --- | --- |
| `.claude/commands/{name}.md` | 단일 슬래시 명령에 적합 |
| `.claude/skills/{name}/SKILL.md` | 보조 파일과 함께 쓰는 고급 스킬에 적합 |

이 장에서는 단순한 문서 갱신 목적이므로 `.claude/commands/update-docs.md`로 충분합니다.

---

# 2.10 2장 가드레일 살펴보기

## 가드레일의 역할

2장에서는 도구 선택보다 실행 순서가 중요합니다.

따라서 `decision-guides`보다 `prompt-guardrails`가 중심이 됩니다.

## 2장에서 사용한 가드레일

| 절 | 가드레일 파일 | 역할 |
| --- | --- | --- |
| 2.2 | `2.2-install-check.md` | 클로드 코드 설치 확인, statusline 설정 |
| 2.3 | `2.3-gcloud.md` | gcloud 설치, 인증, 프로젝트·리전·존 설정 |
| 2.5 | `2.4-github-repo.md` | 저장소 생성, `CLAUDE.md`, 디렉터리 구조 |
| 2.6 | `2.5-gke-cluster.md` | 클러스터 사양, kubeconfig, Gateway API 확인 |
| 2.7 | `2.6-build-deploy.md` | Go 앱, Dockerfile, 매니페스트, 배포 |
| 2.8 | `2.7-first-commit.md` | `JOURNEY.md` 생성, 커밋, 푸시 |

3장부터는 도구를 선택하는 일이 많아지기 때문에 `decision-guides`가 추가됩니다.

예를 들어 배포 자동화 도구를 고를 때 ArgoCD, Flux, Jenkins X 등을 비교한 뒤 선택하게 됩니다.

---

# 2장 핵심 정리

2장의 핵심은 다음과 같습니다.

- GCP 계정을 준비하고 무료 크레딧 범위에서 실습을 시작한다.
- 클로드 코드는 터미널에서 명령을 실행하는 AI 동료 역할을 한다.
- gcloud CLI를 설치하고 GCP 인증을 완료한다.
- 프로젝트, 리전, 존을 기본값으로 설정한다.
- Artifact Registry 인증을 설정해 컨테이너 이미지를 푸시할 수 있게 한다.
- GitHub Copilot과 클로드 코드의 역할 차이를 이해한다.
- GitHub 저장소를 만들고 `CLAUDE.md`로 프로젝트 컨텍스트를 정의한다.
- GKE Standard 클러스터를 Spot VM 기반으로 생성한다.
- kubeconfig를 설정해 `kubectl`이 GKE 클러스터를 바라보게 한다.
- Go 기반 Notiflex API 서버를 작성하고 컨테이너 이미지로 빌드한다.
- Artifact Registry에 이미지를 저장하고 Kubernetes 매니페스트로 배포한다.
- `Namespace`, `Deployment`, `Service`를 통해 앱을 실행한다.
- port-forward로 `/health`, `/id` API를 확인한다.
- `JOURNEY.md`를 생성해 실습 진행 상황을 기록한다.
- 첫 커밋을 GitHub에 푸시하면서 2장을 마무리한다.
- `/update-docs` 스킬로 이후 장의 문서 갱신을 자동화할 준비를 한다.