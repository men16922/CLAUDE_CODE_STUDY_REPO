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
