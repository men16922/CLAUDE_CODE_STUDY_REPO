# Kafka Topic 삭제 절차서

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
