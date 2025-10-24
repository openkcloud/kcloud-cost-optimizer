# OPT CST Controller - Operation Cost Policy Managing Module

**운용 정책 엔진 - 비용 최적화 정책 설정 및 자동화 규칙 관리**

## 📋 주요 기능

### 🎯 정책 관리
- **비용 최적화 정책**: 전력 효율, 비용 제약, 성능 목표 설정
- **자동화 규칙**: 조건-동작 기반 자동화 트리거 정의
- **우선순위 관리**: 워크로드 타입별 우선순위 및 SLA 정책
- **정책 버전 관리**: 정책 변경 이력 및 롤백 지원

### ⚡ 실시간 정책 적용
- **동적 정책 평가**: 워크로드 배치 시 정책 실시간 평가
- **정책 충돌 해결**: 여러 정책 간 충돌 시 우선순위 기반 해결
- **정책 전파**: 모든 모듈에 정책 변경사항 실시간 전파
- **피드백 루프**: 정책 효과 모니터링 및 자동 조정

### 🔄 자동화 엔진
- **이벤트 기반 트리거**: 특정 조건 발생 시 자동 동작 실행
- **시간 기반 스케줄링**: 주기적 정책 실행 (야간 최적화 등)
- **임계값 모니터링**: 비용/전력/성능 임계값 초과 시 알람
- **자동 조치**: 비효율 감지 시 자동 재배치/재구성

## 🏗 아키텍처

```
정책 정의 → Policy Engine → 정책 평가 → 실행 결정
    ↓           ↓              ↓           ↓
  YAML       규칙 엔진     우선순위 해결   optimizer
  JSON       정책 저장소    충돌 감지      core
  API        이벤트 처리    조건 평가      infrastructure
```

## 📜 정책 예시

### **1. 비용 최적화 정책**
```yaml
# policy/cost-optimization.yaml
apiVersion: policy.kcloud.io/v1
kind: CostOptimizationPolicy
metadata:
  name: default-cost-policy
  priority: 100
spec:
  objectives:
    - type: minimize_cost
      weight: 0.7
    - type: maintain_performance  
      weight: 0.3
  
  constraints:
    max_cost_per_hour: 100.0  # $100/hour
    max_power_usage: 5000      # 5000W
    min_efficiency_ratio: 0.7
  
  workload_policies:
    - type: ml_training
      preferred_cluster: gpu_intensive
      max_cost_per_hour: 50.0
      allow_spot_instances: true
    
    - type: inference
      preferred_cluster: npu_optimized
      max_latency_ms: 100
      auto_scale: true
  
  automation:
    - trigger: cluster_utilization < 20%
      action: consolidate_workloads
      delay: 30m
    
    - trigger: power_usage > 4500W
      action: migrate_to_efficient_cluster
      immediate: true
```

### **2. 자동화 규칙**
```yaml
# policy/automation-rules.yaml
apiVersion: policy.kcloud.io/v1
kind: AutomationRule
metadata:
  name: idle-cluster-cleanup
  priority: 50
spec:
  conditions:
    - cluster.utilization < 10
    - cluster.idle_duration > 2h
    - cluster.workload_count == 0
  
  actions:
    - type: notify
      target: operations-team
      message: "Idle cluster detected: {{cluster.name}}"
    
    - type: mark_for_deletion
      grace_period: 1h
    
    - type: delete_cluster
      confirm_with: cost_analysis
  
  exceptions:
    - cluster.labels.persistent == "true"
    - time.hour >= 8 AND time.hour <= 18  # Business hours
```

### **3. 워크로드 우선순위 정책**
```yaml
# policy/workload-priority.yaml
apiVersion: policy.kcloud.io/v1
kind: WorkloadPriorityPolicy
metadata:
  name: workload-priorities
spec:
  priority_classes:
    - name: critical
      value: 1000
      preemptionPolicy: PreemptLowerPriority
      globalDefault: false
      description: "Critical production workloads"
    
    - name: high
      value: 500
      preemptionPolicy: Never
      description: "Important workloads"
    
    - name: normal
      value: 100
      globalDefault: true
      description: "Regular workloads"
    
    - name: low
      value: 10
      preemptionPolicy: Never
      description: "Best-effort workloads"
  
  workload_mapping:
    - pattern: "prod-*"
      priority_class: critical
    - pattern: "inference-*"
      priority_class: high
    - pattern: "training-*"
      priority_class: normal
    - pattern: "test-*"
      priority_class: low
```

## 🔧 Go 구현 구조

```go
// policy/internal/engine/engine.go
type PolicyEngine struct {
    rules       RuleStore
    evaluator   PolicyEvaluator
    enforcer    PolicyEnforcer
    notifier    EventNotifier
}

func (e *PolicyEngine) EvaluateWorkload(workload *Workload) (*Decision, error) {
    // 1. 적용 가능한 정책 찾기
    policies := e.rules.GetApplicablePolicies(workload)
    
    // 2. 정책 평가
    results := e.evaluator.Evaluate(workload, policies)
    
    // 3. 충돌 해결
    decision := e.resolveConflicts(results)
    
    // 4. 결정 실행
    e.enforcer.Enforce(decision)
    
    // 5. 이벤트 발생
    e.notifier.Notify(PolicyAppliedEvent{
        Workload: workload,
        Decision: decision,
    })
    
    return decision, nil
}
```

## 📊 API 엔드포인트

```bash
# 정책 관리
GET    /policies                     # 모든 정책 목록
POST   /policies                     # 새 정책 생성
GET    /policies/{policy_id}         # 정책 상세 조회
PUT    /policies/{policy_id}         # 정책 수정
DELETE /policies/{policy_id}         # 정책 삭제
POST   /policies/{policy_id}/enable  # 정책 활성화
POST   /policies/{policy_id}/disable # 정책 비활성화

# 정책 평가
POST   /evaluate/workload            # 워크로드에 대한 정책 평가
POST   /evaluate/cluster             # 클러스터 정책 평가
GET    /evaluate/conflicts           # 정책 충돌 확인

# 자동화 규칙
GET    /rules                        # 자동화 규칙 목록
POST   /rules                        # 규칙 생성
PUT    /rules/{rule_id}             # 규칙 수정
DELETE /rules/{rule_id}             # 규칙 삭제
GET    /rules/{rule_id}/history     # 규칙 실행 이력

# 정책 효과 분석
GET    /analytics/policy-impact      # 정책 영향 분석
GET    /analytics/cost-savings      # 비용 절감 효과
GET    /analytics/compliance        # 정책 준수율
```

## 🧪 사용 예시

### **Go 클라이언트**
```go
package main

import (
    "github.com/kcloud-opt/policy/client"
    "github.com/kcloud-opt/policy/types"
)

func main() {
    // Policy 클라이언트 초기화
    policyClient := client.NewPolicyClient("http://localhost:8005")
    
    // 비용 최적화 정책 생성
    policy := &types.CostOptimizationPolicy{
        Name: "aggressive-cost-saving",
        Objectives: []types.Objective{
            {Type: "minimize_cost", Weight: 0.9},
            {Type: "maintain_performance", Weight: 0.1},
        },
        Constraints: types.Constraints{
            MaxCostPerHour: 80.0,
            MaxPowerUsage:  4000,
        },
    }
    
    // 정책 적용
    err := policyClient.CreatePolicy(policy)
    if err != nil {
        log.Fatal(err)
    }
    
    // 워크로드 평가
    decision, err := policyClient.EvaluateWorkload(&types.Workload{
        ID:   "ml-training-123",
        Type: "ml_training",
        Requirements: types.Requirements{
            CPU:    16,
            Memory: "64Gi",
            GPU:    4,
        },
    })
    
    fmt.Printf("추천 클러스터: %s\n", decision.RecommendedCluster)
    fmt.Printf("예상 비용: $%.2f/hour\n", decision.EstimatedCost)
}
```

### **정책 YAML 적용**
```bash
# 정책 파일 적용
kubectl apply -f policies/cost-optimization.yaml
kubectl apply -f policies/automation-rules.yaml

# 또는 API로 직접 적용
curl -X POST http://localhost:8005/policies \
  -H "Content-Type: application/yaml" \
  -d @policies/cost-optimization.yaml

# 정책 상태 확인
kubectl get policies
kubectl describe policy default-cost-policy
```

## 🔒 보안 설정 (중요!)

**⚠️ 이 저장소는 프로덕션 비밀번호를 포함하지 않습니다.**

### 배포 전 필수 작업

1. **환경 변수 설정**
   ```bash
   # .env.example을 .env로 복사
   cp .env.example .env

   # .env 파일을 편집하여 강력한 비밀번호로 변경
   vi .env
   ```

2. **강력한 비밀번호 생성**
   ```bash
   # 랜덤 비밀번호 생성 (32자)
   openssl rand -base64 32

   # 또는
   pwgen -s 32 1
   ```

3. **필수 환경 변수**
   - `POSTGRES_PASSWORD`: PostgreSQL 데이터베이스 비밀번호
   - `REDIS_PASSWORD`: Redis 캐시 비밀번호
   - `GRAFANA_ADMIN_PASSWORD`: Grafana 관리자 비밀번호
   - `API_TOKEN`: API 인증 토큰
   - `WEBHOOK_TOKEN`: Webhook 인증 토큰

4. **Kubernetes 비밀 설정**
   ```bash
   # secret.yaml.example을 secret.yaml로 복사
   cp k8s/secret.yaml.example k8s/secret.yaml

   # 실제 값으로 수정 후 적용
   kubectl apply -f k8s/secret.yaml

   # 보안 강화: Sealed Secrets 사용 권장
   # https://github.com/bitnami-labs/sealed-secrets
   ```

### 보안 권장 사항

- ✅ **절대 .env 파일을 커밋하지 마세요** (.gitignore에 포함됨)
- ✅ **k8s/secret.yaml을 커밋하지 마세요** (.gitignore에 포함됨)
- ✅ 최소 16자 이상의 복잡한 비밀번호 사용
- ✅ 프로덕션 환경에서는 비밀 관리 도구 사용:
  - Kubernetes: Sealed Secrets, External Secrets Operator
  - Cloud: AWS Secrets Manager, Azure Key Vault, GCP Secret Manager
  - Self-hosted: HashiCorp Vault
- ✅ 정기적인 비밀번호 로테이션
- ✅ 네트워크 방화벽 설정 (필요한 포트만 개방)

### 프로덕션 배포 체크리스트

- [ ] 모든 기본 비밀번호를 강력한 비밀번호로 변경
- [ ] .env 파일 생성 및 설정
- [ ] Kubernetes Secrets 설정
- [ ] HTTPS/TLS 인증서 설정
- [ ] 방화벽 및 네트워크 정책 구성
- [ ] 모니터링 및 로깅 설정
- [ ] 백업 전략 수립

## 🚀 배포

```bash
# 로컬 개발
make build
make test
make run

# Docker 빌드 및 실행
make docker-build
make docker-run

# K8s 배포
kubectl apply -f deployment/policy.yaml

# 정책 초기화
make init-policies
```

## 📈 요구사항 충족

- **SFR.OPT.024**: 플랫폼 운용 비용 최적화 정책 설정/관리 ✅
- **SFR.OPT.030**: 자동화 정책 정의 기능 ✅
- **정책 기반 의사결정**: 모든 스케줄링/재배치 결정에 정책 적용
- **실시간 정책 업데이트**: 재시작 없이 정책 변경 가능
