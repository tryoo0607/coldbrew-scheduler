package clientgo

import "k8s.io/client-go/kubernetes"

// Fake 테스트용
// Test파일에서 FakeClient를 Custom 하기 때문에 Custom한 cs를 그대로 client로 변경 필요함
// 때문에 아래의 메소드를 별도로 생성
func NewWithClientset(cs kubernetes.Interface) Client {
    return newClient(cs)
}