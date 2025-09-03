
# 📦 clientgo 패키지 구조 및 설계 가이드
`clientgo`는 **Kubernetes client-go 의존성을 내부에 격리**하고,  
바깥 레이어에서는 **도메인 모델(api 패키지)** 만을 사용하도록 설계된 패키지입니다.



## 🔗 의존성 흐름 (Dependency Flow)

```


                 ┌────────────┐
                 │   clientgo │  ← 외부 파사드
                 └────┬───────┘
                      │
         ┌────────────▼────────────┐
         │         k8s             │  ← 내부 전용 (ClientSet 생성/감지)
         └────────────┬────────────┘
                      │
        ┌─────────────▼─────────────┐
        │        informer           │  ← Pod 감지, 콜백 트리거
        └─────┬─────────────┬───────┘
              │             │
              │             │
        ┌─────▼────┐   ┌────▼──────┐
        │ adaptor  │   │  binder   │
        └─────┬────┘   └────▲──────┘
              │             │
              ▼             │
         ┌────────┐         │
         │  api   │◀────────┘
         └────────┘
```

```
app (project, scheduler, finder)
  |
  v
clientgo (Facade)
  |---> k8s       // clientset 생성 (InCluster/Kubeconfig/Fake)
  |---> informer  // Pod 인포머 구성 (kubernetes.Interface만 의존)
  |---> adapter   // k8s → api 변환
        |
        v
       api        // 도메인 타입/콜백 (k8s import 금지)
        ^
        |
      binder      // Binding 실행 (subresource → 검증 → 폴백)
```



---

## 📂 api/

- `types.go`  
  - `PodInfo`, `Strategy`/콜백 시그니처 등 **도메인 모델** 정의
- `constants.go`  
  - 필드 셀렉터 키, 스케줄러 이름 등 **공용 상수**
- `errors.go`  
  - 선택적 도메인 오류 타입 정의

### 원칙

- 여기서는 `k8s` 타입을 절대 import 하지 않습니다.
- 도메인 모델(`PodInfo`)은 Kubernetes 의존성과 분리되어  
  스케줄러 정책 로직이 **k8s를 몰라도 동작**하도록 보장합니다.

---

## 📂 adaptor/

- `podinfo.go`
  - `*corev1.Pod` → `api.PodInfo` 변환
  - 리소스 요청 합계(CPU milli, Memory bytes) 계산 포함

### 특징

- 변환 로직은 **순수 함수**로 유지 → 단위 테스트 용이
- k8s 리소스를 도메인 모델로 **안전하게 변환**하는 역할만 담당

---

## 📂 informer/

- `listwatcher.go`
  - `ListWatch` 생성 (필드/라벨 셀렉터 조립)
- `podinformer.go`
  - `NewPodInformer(cs, findBestNode)` 생성
  - Add 이벤트 핸들러에서 동작 순서:
    1. `adaptor.ToPodInfo(pod)`
    2. `findBestNode(podInfo)` 호출
    3. `binder`를 통해 노드에 Pod 바인딩

---

## 📂 binder/

- `binder.go`
  - `BindPodToNode(opts)` 또는 `BindToNodeWithBinding(opts)` 제공

- BindOptions 구조체:

  - `ClientSet`: kubernetes.Interface  
  - `Ctx`: context.Context  
  - `Pod`: *corev1.Pod  
  - `NodeName`: string

### 원칙

- `api` 패키지 import ❌  
- 실클러스터: Subresource 바인딩 우선 시도  
- FakeClient: 서브리소스 호출은 무시됨 → `spec.nodeName` 직접 업데이트  

---

## 📂 k8s/ (옵션)

- `client.go`
  - `kubernetes.Interface` 생성/설정 도우미
  - `out-of-tree auth`, `kubeconfig` 설정 등 클라이언트 초기화 기능 제공
  - `NewFakeClientset()`도 여기에 포함

- `IsFakeClient()`  
  - 내부적으로 `client-go/testing.Fake` 타입인지 확인

### 원칙

- 이 패키지는 내부 전용 (clientgo 내부에서만 사용)
- 외부에서는 직접 import ❌

---

## 📂 clientgo (파사드)

- 외부에 노출되는 인터페이스: `Client`

  - `ListNodeInfos(ctx)`
  - `NewPodInformer(ctx, findFunc)`

- 내부 생성자들:

  - `New(opt Options)`: 클러스터 종류에 따라 ClientSet 생성
  - `NewWithClientset(cs)`: 테스트 목적의 주입용
  - 내부적으로는 모두 `newClient(cs)` 호출

- `options.go`

  - `Options` 구조체
    - `InCluster`: bool
    - `Kubeconfig`: string
    - `UseFake`: bool

---

## ❗️왜 BindOptions는 api가 아니라 binder에 있어야 하나?

### api는 도메인 모델 전용 패키지

- `PodInfo` 같은 **스케줄러 정책에서 쓰는 정보 모델**만 존재
- `*corev1.Pod`, `ClientSet` 등을 알면 안 됨
- 만약 `api.BindOptions`에 k8s 타입을 넣으면:
  - `api` 패키지가 Kubernetes 의존성에 종속됨 → **DIP (의존성 역전 원칙) 위반**

### binder는 실제 바인딩 책임 패키지

- Kubernetes API 서버와 통신하는 역할
- `ClientSet`, `Pod`, `NodeName` 등 **실행 파라미터**를 모아둠
- 따라서 `BindOptions`는 `binder`에 정의하는 것이 자연스러움

### 요약 표

| 항목            | 적절한 위치 | k8s 의존성 | 역할                                         |
|-----------------|--------------|-------------|----------------------------------------------|
| `PodInfo`       | `api`        | ❌ 없음      | 스케줄러 정책에서 사용하는 도메인 모델        |
| `BindOptions`   | `binder`     | ✅ 있음      | Kubernetes API 호출을 위한 실행 파라미터      |

---

## 핵심 원칙 ✅

- `client-go` 호출은 **항상 clientgo 내부에서만!**
- `api` 패키지는 **k8s 타입을 몰라야 함**
- `adaptor`는 **k8s → 도메인 모델 변환**만 담당
- `informer`는 이벤트 핸들링 및 스케줄링 정책 호출 조립
- `binder`는 **바인딩 실행 로직**만 담당

