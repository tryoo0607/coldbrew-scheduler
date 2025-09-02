# 의존성 흐름
```
api            (도메인 타입만; k8s import 금지)
 ↑
adaptor        (k8s → api 변환)
 ↑        ↘
informer  →  binder
   ↑
  k8s
```


# clientgo 패키지 구조 및 역할

`clientgo`는 **Kubernetes client-go 의존성을 내부에 격리**하고,  
바깥 레이어에서는 **도메인 모델(api 패키지)** 만을 사용하도록 설계된 패키지입니다.


---

## 📂 api/

- **`types.go`**  
  - `PodInfo`, `Strategy`/콜백 시그니처 등 **도메인 모델** 정의
- **`constants.go`**  
  - 필드 셀렉터 키, 스케줄러 이름 등 **공용 상수**
- **`errors.go`**  
  - 선택적 도메인 오류 타입 정의

### **원칙**
- **여기서는 `k8s` 타입을 절대 import 하지 않습니다.**
- 도메인 모델(`PodInfo`)은 Kubernetes 의존성과 분리하여  
  스케줄러 정책 로직이 **k8s를 몰라도 동작**하도록 보장합니다.

---

## 📂 adaptor/

- **`podinfo.go`**
  - `*corev1.Pod` → `api.PodInfo` 변환
  - 리소스 요청 합계(CPU milli, Memory bytes) 계산 포함

### **특징**
- 변환 로직은 **순수 함수**로 유지 → 단위 테스트 용이
- k8s 리소스를 도메인 모델로 **안전하게 변환**하는 역할만 담당

---

## 📂 informer/

- **`listwatcher.go`**
  - `ListWatch` 생성 (필드/라벨 셀렉터 조립)
- **`podinformer.go`**
  - `NewPodInformer(cs, findBestNode)` 생성
  - Add 이벤트 핸들러에서 동작 순서:
    1. `adaptor.ToPodInfo(pod)`
    2. `findBestNode(podInfo)` 호출
    3. `binder`를 통해 노드에 Pod 바인딩

---

## 📂 binder/

- **`binder.go`**
  - `BindPodToNode(opts)` 또는 `BindToNodeWithBinding(opts)` 제공
  - **`Options`**:
    - `ClientSet`
    - `Ctx`
    - `Pod`
    - `NodeName`
  - 실제 Kubernetes **Binding 서브리소스** 또는 **spec.nodeName 패치** 호출을 **캡슐화**

---

## 📂 k8s/ *(옵션)*

- **`client.go`**
  - `kubernetes.Interface` 생성/설정 도우미
  - `out-of-tree auth`, `kubeconfig` 설정 등 클라이언트 초기화 기능 제공

---

## ❗️왜 `BindOptions`는 `api`가 아니라 `binder`에 있어야 하나?

### `api`는 **도메인 모델 전용 패키지**
- `PodInfo` 같은 **스케줄러 정책에서 쓰는 정보 모델**만 존재.
- **k8s 타입(`*corev1.Pod`, `ClientSet`)을 전혀 몰라야**  
  스케줄러 정책 코드가 Kubernetes 의존성과 분리됩니다.
- 만약 `api.BindOptions`에 `*corev1.Pod`나 `ClientSet`을 넣으면,  
  **api 패키지가 k8s에 종속 → 의존성 역전 원칙(DIP) 깨짐**.

### `binder`는 **실제 바인딩 실행 책임 패키지**
- `binder`는 Kubernetes API 서버와 통신해 Pod를 특정 노드에 스케줄링하는 역할.
- 따라서 `ClientSet`, `Pod`, `NodeName` 등 **실행 파라미터**를  
  `binder.Options`(=BindOptions)로 모아두는 게 자연스럽습니다.

### 요약
| 항목          | 적절한 위치 | k8s 의존성 | 역할 |
|---------------|------------|-----------|------|
| **PodInfo**   | `api`      | ❌ 없음   | 스케줄러 정책에서 사용하는 도메인 모델 |
| **BindOptions** | `binder` | ✅ 있음   | Kubernetes API 호출을 위한 실행 파라미터 |

---

## 핵심 원칙 ✅

- **`client-go` 호출은 clientgo 내부에서만!**  
- **`api` 패키지는 k8s 타입을 몰라야 함.**  
- `adaptor`는 **k8s → 도메인 모델 변환**만 담당  
- `informer`는 이벤트 핸들링 및 스케줄링 정책 호출까지 조립  
- `binder`는 **실제 스케줄링 반영** 로직만 담당  