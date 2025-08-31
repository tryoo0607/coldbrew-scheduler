## Project 구조
- Golang 권장하는, standard 프로젝트 구조에 대해 적어둔 Github을 찾음
- 해당 구조를 참고하여 작업 중

## Config 고민
- config는 viper로 구성하기
- viper는 env 파일 불러오는데 용이, EnvConfig를 불러오기는 어려움
- 별도 라이브러리인 envconfig 함께 사용하도록 결정
- envconfig -> 환경변수 읽어오기, viper -> App 설정 읽어오도록 역할 양분

- 환경변수를 문서화 해야하는데 누락하면?
- envconfig 라이브러리에서 제공해주는 함수를 통해 환경변수 목록 가져올 수 있음
- 해당 내용 출력하는 별도 패키지를 `cmd`에 패키지에 구성

- config 값을 사용하는 방법
    - main에서 내부로 전달하는 방식
    - import를 통해 가져오는 방식

    - import를 하게되면 결합도가 높아지게 되므로 main에서 전달하는 방식이 나쁘지 않아보임
    - 일단은 이렇게 진행


## Model 설정
- 매개변수가 많은 경우
    - 자식쪽에서 struct 생성하기
    - 어차피 자식 쪽 메서드를 사용하려면 import 하기 때문에

- 공통으로 사용되는 경우가 많다면 
    - internal이든 어디든 별도 model 패키지로 뺴기

## Error Handling 고민
- Error Handling 고민 중
- Golang에서 Error는 보통 Caller가 다룬다고 함
    - 그런데 이걸 어디까지 올려야하지? Main까지?
    - 최종적으로 Main에서 Error를 다루도록 하는게 맞는건가?
    - 아니면 고루틴까지?
    - => Caller가 다룬 다는 것은 최상위까지 올리라는 의미가 아니고 말 그대로 호출한 곳에서 error 처리를 한다는 의미
    
    - 그런데 Context에 따라 Wrapping하는 것이 좋다고함
    - 그럼 언제 Wrapping을 해야하는지?
    - 
- 자주 사용하는 것들은 Custom Error를 만들라고하는데 
    - Error 패키지를 별도로 두는 것이 좋은지? 아니면 해당 패키지 내에서 사용하는 것이 좋은지
    - Error 패키지를 만들게되면 Error 패키지에 의존성이 생기므로 별로 좋은 방법 같지는 않음
    - 패키지 내에서 별도로 model이나 error만을 다루는 파일이 필요해보임
    - 그런데 공통된 에러는 InvalidInput 이런거라던가

## Logging 처리 고민
- 공통적으로 Zap이란 라이브러리를 많이 추천함 => 사용해보기
- 