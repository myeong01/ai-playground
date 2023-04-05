# ai-playground

AI 분석 환경을 쉽게 구축하고 관리하기 위한 프로젝트로 Kubernetes Custom Resource Controller 를 통하여 구현하였습니다.

### 컴포넌트

```
└ cmd
  └ api         // 각종 Custom Resource 에 대한 CRUD API 컴포넌트
  └ authz       // CRUD API 호출에 대한 유저의 권한 확인 (K8S 의 인가 시스템을 이용하며 SubjectAccessReviewSpec 를 통하여 확인)
  └ controllers // 각종 Custom Resource 에 대한 Controller
```

### Custom Resource

```
- ClusterRole        // 공개 O, 권한 허가, 부모 ClusterRole 상속 등의 추가 기능이 필요하여 별도의 ClusterRole 정의
- Group              // 공개 O, 유저가 속한 그룹 정의
- Role               // 공개 O, 권한 허가, 부모 Role 상속 등의 추가 기능이 필요하여 별도의 Role 정의
- Container          // 공개 O, Jupyter 서비스를 띄우기 위한 Deployment, Service, VirtualService 를 관리하는 리소스
- ContainerSnapshot  // 공개 X, Pod 의 컨테이너를 버저닝
- Dataset            // 공개 O, 다이나믹 PVC 마운트 기능을 위해 PVC 를 래핑하는 별도의 Dataset 정의
- DynamicMount       // 공개 X, 동작 중인 Pod 의 중단 없이 데이터셋을 마운트하는 리소스
- Image              // 공개 O, Container 에서 배포할 때 사용할 컨테이너 이미지 정의
- Resource           // 공개 O, Container 에서 배포할 때 사용할 ResourceQuota 템플릿
- Playground         // 공개 O, MultiTenant 의 단위로 Kubernetes 의 namespace 와 1 대 1 대응
```
