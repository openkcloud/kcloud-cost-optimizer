# kcloud-opt-optimizer

AI반도체 클라우드 비용 최적화 엔진 (Go)

## 개요

kcloud-opt-optimizer는 AI반도체 워크로드의 운영 비용을 분석하고 최적화하는 고성능 엔진입니다. 실시간 비용 분석, 인프라 재구성 권장, 그리고 성능-비용 트레이드오프 최적화를 통해 클라우드 운영 효율성을 극대화합니다.

## 주요 기능

### 비용 최적화 분석
- **실시간 비용 추적**: 워크로드별 실시간 운영 비용 모니터링
- **효율성 분석**: 비용 대비 성능 효율성 계산 및 분석
- **트레이드오프 최적화**: 성능과 비용 간의 최적 균형점 도출

### 인프라 재구성
- **클러스터 재구성**: 비용 효율적인 가상 클러스터 재배치
- **자원 통합**: 유휴 자원 최적화를 통한 비용 절감
- **스팟 인스턴스 활용**: 배치 작업의 스팟 인스턴스 자동 마이그레이션

### 정책 기반 최적화
- **에너지 정책**: 전력 효율성 기반 최적화 정책
- **비용 정책**: 예산 제약 조건 하에서의 최적화
- **SLA 보장**: 성능 SLA를 유지하면서 비용 최소화

## 아키텍처

```
optimizer/
├── cmd/optimizer/           # 메인 애플리케이션
├── pkg/
│   ├── cost_optimizer/     # 비용 최적화 로직
│   ├── infrastructure/     # 인프라 재구성
│   ├── policies/           # 정책 엔진
│   └── algorithms/         # 최적화 알고리즘
├── internal/
│   ├── models/            # 데이터 모델
│   └── services/          # 비즈니스 서비스
├── config/                # 설정 파일
└── tests/                # 테스트
```

## 설치 및 실행

### 개발 환경

```bash
# 의존성 다운로드
make deps

# 빌드
make build

# 테스트
make test

# 로컬 실행
make run
```

### Docker

```bash
# 이미지 빌드
make docker-build

# 컨테이너 실행
docker run -p 8003:8003 kcloud-opt/optimizer:latest
```

### Kubernetes

```bash
# ConfigMap 적용
kubectl apply -f config/configmap.yaml

# 배포
kubectl apply -f config/deployment.yaml
```

## 설정

### 환경변수

- `CORE_API`: Core Scheduler API URL
- `ANALYZER_API`: Analyzer Service API URL
- `REDIS_HOST`: Redis 캐시 서버
- `OPTIMIZATION_INTERVAL`: 최적화 실행 주기 (초)

### 설정 파일 (config.yaml)

```yaml
optimizer:
  # 최적화 정책
  policies:
    cost_optimization:
      enabled: true
      target_reduction: 20  # 목표 비용 절감율 (%)
      max_performance_impact: 5  # 최대 성능 영향도 (%)
    
    energy_efficiency:
      enabled: true
      target_power_reduction: 15
      prefer_green_energy: true
    
    sla_constraints:
      latency_threshold: 100  # ms
      availability_threshold: 99.9  # %
      
  # 알고리즘 설정
  algorithms:
    genetic_algorithm:
      population_size: 100
      generations: 50
      mutation_rate: 0.1
    
    simulated_annealing:
      initial_temperature: 1000
      cooling_rate: 0.95
      min_temperature: 1
```

## API 엔드포인트

### 최적화 관리
- `GET /health`: 서비스 상태 확인
- `POST /api/v1/optimize`: 최적화 실행
- `GET /api/v1/optimize/status/{job_id}`: 최적화 작업 상태 조회
- `GET /api/v1/optimize/results/{job_id}`: 최적화 결과 조회

### 비용 분석
- `GET /api/v1/cost/analysis`: 현재 비용 분석 결과
- `GET /api/v1/cost/trends`: 비용 변화 추이
- `POST /api/v1/cost/forecast`: 비용 예측 요청

### 인프라 관리
- `GET /api/v1/infrastructure/efficiency`: 인프라 효율성 분석
- `POST /api/v1/infrastructure/reconfig`: 인프라 재구성 요청
- `GET /api/v1/infrastructure/recommendations`: 최적화 권장사항

## 최적화 알고리즘

### 1. 유전 알고리즘 (Genetic Algorithm)

```go
type GeneticOptimizer struct {
    PopulationSize int
    Generations    int
    MutationRate   float64
}

func (g *GeneticOptimizer) Optimize(workloads []Workload, nodes []Node) (*OptimizationResult, error) {
    // 초기 개체군 생성
    population := g.initializePopulation(workloads, nodes)
    
    for generation := 0; generation < g.Generations; generation++ {
        // 적합도 평가
        fitness := g.evaluateFitness(population)
        
        // 선택, 교배, 돌연변이
        newGeneration := g.evolve(population, fitness)
        population = newGeneration
    }
    
    return g.getBestSolution(population), nil
}
```

### 2. 담금질 기법 (Simulated Annealing)

```go
type SimulatedAnnealingOptimizer struct {
    InitialTemp float64
    CoolingRate float64
    MinTemp     float64
}

func (sa *SimulatedAnnealingOptimizer) Optimize(initial *Solution) (*Solution, error) {
    current := initial
    best := current.Copy()
    temperature := sa.InitialTemp
    
    for temperature > sa.MinTemp {
        neighbor := sa.generateNeighbor(current)
        deltaE := neighbor.Cost - current.Cost
        
        if deltaE < 0 || math.Exp(-deltaE/temperature) > rand.Float64() {
            current = neighbor
            if current.Cost < best.Cost {
                best = current.Copy()
            }
        }
        
        temperature *= sa.CoolingRate
    }
    
    return best, nil
}
```

### 3. 정수 선형 계획법 (ILP)

```go
type ILPOptimizer struct {
    Solver string // "gurobi", "cplex", "glpk"
}

func (ilp *ILPOptimizer) Optimize(constraints *OptimizationConstraints) (*Solution, error) {
    model := ilp.buildModel(constraints)
    
    // 목적 함수: 총 비용 최소화
    model.SetObjective("minimize total_cost")
    
    // 제약 조건
    model.AddConstraint("resource_capacity", "sum(allocated_resources) <= total_capacity")
    model.AddConstraint("sla_constraint", "performance_score >= min_performance")
    model.AddConstraint("power_constraint", "total_power <= power_budget")
    
    solution, err := model.Solve()
    return solution, err
}
```

## 최적화 전략

### 비용-성능 트레이드오프

```go
type TradeoffOptimizer struct {
    CostWeight        float64  // 비용 가중치
    PerformanceWeight float64  // 성능 가중치
    PowerWeight       float64  // 전력 가중치
}

func (t *TradeoffOptimizer) CalculateScore(solution *Solution) float64 {
    // 정규화된 점수 계산
    costScore := 1.0 - (solution.Cost / t.MaxCost)
    perfScore := solution.Performance / t.MaxPerformance
    powerScore := 1.0 - (solution.PowerUsage / t.MaxPower)
    
    return t.CostWeight*costScore + 
           t.PerformanceWeight*perfScore + 
           t.PowerWeight*powerScore
}
```

### 동적 임계값 조정

```go
func (o *Optimizer) AdjustThresholds(historicalData *HistoricalMetrics) {
    // 과거 성능 데이터 기반 임계값 동적 조정
    avgLatency := historicalData.AverageLatency()
    if avgLatency < o.LatencyThreshold*0.8 {
        // 성능에 여유가 있으면 비용 최적화 강화
        o.CostWeight += 0.1
        o.PerformanceWeight -= 0.1
    } else if avgLatency > o.LatencyThreshold*0.95 {
        // 성능이 임계점에 가까우면 성능 우선
        o.PerformanceWeight += 0.1
        o.CostWeight -= 0.1
    }
}
```

## 사용 예시

### 1. 비용 최적화 실행

```bash
# 최적화 작업 시작
curl -X POST http://localhost:8003/api/v1/optimize \
  -H "Content-Type: application/json" \
  -d '{
    "workloads": ["training-job-*"],
    "optimization_goals": {
      "cost_reduction": 20,
      "max_performance_impact": 5
    },
    "algorithm": "genetic"
  }'

# 결과: {"job_id": "opt-12345", "status": "running"}
```

### 2. 최적화 결과 조회

```bash
curl http://localhost:8003/api/v1/optimize/results/opt-12345

# 응답
{
  "job_id": "opt-12345",
  "status": "completed",
  "results": {
    "cost_reduction": 18.5,
    "performance_impact": 2.1,
    "power_savings": 250.0,
    "recommendations": [
      {
        "type": "migrate_workload",
        "workload": "training-job-001",
        "from": "gpu-node-01",
        "to": "gpu-node-03",
        "savings": 8.50
      }
    ]
  }
}
```

### 3. 인프라 재구성 권장사항

```bash
curl http://localhost:8003/api/v1/infrastructure/recommendations

{
  "recommendations": [
    {
      "type": "consolidate_workloads",
      "description": "Consolidate 3 small workloads to 1 larger node",
      "estimated_savings": 45.20,
      "confidence": 0.92
    },
    {
      "type": "spot_instance_migration", 
      "description": "Migrate batch jobs to spot instances",
      "estimated_savings": 67.80,
      "confidence": 0.87
    }
  ]
}
```

## 모니터링

### Prometheus 메트릭

- `kcloud_optimizer_cost_savings_total`: 누적 비용 절약
- `kcloud_optimizer_optimizations_total`: 최적화 실행 횟수
- `kcloud_optimizer_efficiency_score`: 현재 효율성 점수
- `kcloud_optimizer_algorithm_duration`: 알고리즘 실행 시간

### 대시보드

- **비용 절약 대시보드**: 실시간 비용 절약 현황
- **효율성 분석**: 자원 활용 효율성 추이
- **최적화 히스토리**: 과거 최적화 작업 결과

## 개발

### 요구사항

- Go 1.21+
- Redis 6+
- PostgreSQL 12+ (메타데이터 저장)

### 테스트

```bash
# 단위 테스트
make test

# 벤치마크 테스트
make benchmark

# 알고리즘 성능 테스트
go test -run=TestAlgorithmPerformance -v
```

### 알고리즘 추가

```go
// 새로운 최적화 알고리즘 인터페이스
type OptimizationAlgorithm interface {
    Name() string
    Optimize(workloads []Workload, constraints *Constraints) (*Solution, error)
    Validate(solution *Solution) error
}

// 알고리즘 등록
func RegisterAlgorithm(name string, algorithm OptimizationAlgorithm) {
    algorithmRegistry[name] = algorithm
}
```

## 라이선스

Apache License 2.0