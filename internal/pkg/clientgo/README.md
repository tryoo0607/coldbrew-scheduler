
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
        │        informer           │  ← 객체 감지, Pod 이벤트 핸들러
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

- `nodeinfo.go`
  - `*corev1.Node` → `api.NodeInfo` 변환
  - 리스트 변환 포함

### 특징

- 변환 로직은 **순수 함수**로 유지 → 단위 테스트 용이
- k8s 리소스를 도메인 모델로 **안전하게 변환**하는 역할만 담당

---

## 📂 informer/

- `factory.go`
  - `NewInformerFactory(cs)` 사용 해서 동시에 Pod, Node 조회를 가능

- `podinformer.go`

  - `NewPodController(ctx, cs, podInformer, nodeLister, findFunc)`
  - Pod Add 이벤트에서 동작 순서:

    1. `ToPodInfo(pod)`
    2. `ListNodeInfos()`
    3. `ToNodeInfoList()`
    4. `findBestNode(podInfo, nodes)`
    5. `binder.Bind(...)`

- `nodeinformer.go`

  - `NodeController` 객체와 Lister 구성


### 원칙

- `kubernetes.Interface`만 의존하며, `client-go`의 구체 타입은 알지 못함.

- 모든 k8s 객체는 이벤트 수신 후 `adapter`를 통해 도메인 모델로 변환해서 외부(`scheduler`, `finder`)에 넘겨줌.

- 외부에는 `Run()` 메서드만 제공하는 `Controller` 인터페이스 형태로 노출됨.

- informer 내부에서만 client-go의 Informer, Lister 사용 → 외부(`scheduler`, c`lientgo`)에서는 도메인 모델만 다룸.
---

## 📂 binder/

- `binder.go`
  - `BindPodToNode(opts)` 또는 `BindToNodeWithBinding(opts)` 제공

- `BindOptions` 구조체:

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

- `clientset.go`
  - InCluster, Kubeconfig, FakeClient에 따라 kubernetes.Interface 생성

- `helper.go`
  - `resolveKubeconfigPath()` CLI 인자, 환경변수, 기본 경로 (~/.kube/config) 우선순위로 kubeconfig 경로 결정
  - `isFakeClient()` 주어진 Interface가 *fake.Clientset 인지 여부 판단


### 원칙

- 이 패키지는 내부 전용입니다. `clientgo` 내부에서만 import 해야 하며, 외부에서는 직접 사용하지 않습니다.

- 목적은 `clientset` 생성/감지 분리로, 클러스터 타입에 따른 초기화를 담당하며 나머지 로직에서 클러스터 환경을 의식하지 않도록 합니다.

- 테스트에서는 `NewFakeClientset()` 을 통해 실제 클러스터 없이도 로직 테스트가 가능하도록 지원합니다.

---

## 📂 clientgo (파사드)

### 환경

* `Options` 출입: `Kubeconfig`, `InCluster`, `UseFake`

### 구성 가정

```go
func New(opt Options) (Client, error) {
    return newClient(cs), nil
}
```

### 개발자에게 보여주는 Facade 인터페이스

```go
type Client interface {
    ListNodeInfos(ctx context.Context) ([]api.NodeInfo, error)
    NewPodController(ctx context.Context, find api.FinderFunc) (Controller, error)
}
```

### 입력을 받는 구현체

```go
type client struct {
    cs kubernetes.Interface
}

func (c *client) NewPodController(...) (Controller, error) {
    factory := informer.NewInformerFactory(c.cs)
    podInformer := factory.Core().V1().Pods()
    nodeInformer := factory.Core().V1().Nodes()

    return informer.NewPodController(
        ctx, c.cs, podInformer, nodeInformer.Lister(), find
    ), nil
}
```

### 외부에 보여주는 Controller 객체 프린 인터페이스

```go
type Controller interface {
    Run(stopCh <-chan struct{})
}
```


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


| 구조         | 개요             | k8s 타입 의심 | 속성                   |
| ---------- | -------------- | --------- | -------------------- |
| `api`      | 도메인 타입         | ❌ 제한해야함   | 양의, 디콜로드 가능          |
| `adapter`  | k8s → 도메인      | ✅ 사용      | 해당성이 매우 고유           |
| `informer` | Pod/Node Event | ✅ 의심      | policy handler 연결    |
| `binder`   | 바인딩 진행         | ✅ 반복 검증까지 | Pod Binding          |
| `k8s`      | clientset 생성   | ✅ 만적      | internal-only        |
| `clientgo` | 파사드            | ❌ 보복물 무   | 외부가 k8s 모드 알해야할 필요 X |

---


## 테스트
```bash
go test ./internal/app/scheduler -v -count=1
```