## To build and test your code you need to:

1. Download and install local k8s environment, such as Minikube or [Kind](https://kind.sigs.k8s.io/docs/user/quick-start/).
2. Fill the .env file.
3. Run IMG="your image tag" **make docker-build**
4. Run IMG="your image tag" **make docker-push**
5. Run IMG="your image tag" **make deploy**
6. Apply the yaml manifest from operator/config/samples
7. Check logs of the operator container **kubectl logs -n operator-system operator-controller-manager-xxx**
8. Fix the issue if something goes wrong and repeat.