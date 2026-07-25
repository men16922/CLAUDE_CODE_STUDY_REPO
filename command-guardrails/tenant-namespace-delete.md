# Tenant Namespace 삭제 절차서

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
