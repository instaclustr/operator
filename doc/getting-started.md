## To build and test your code you need to:

1. Download and install local k8s environment, such as Minikube or [Kind](https://kind.sigs.k8s.io/docs/user/quick-start/).
2. Fill the .env file with your credentials.
3. Run **make cert-deploy**
4. Run IMG="your image tag" **make docker-build**
5. Run IMG="your image tag" **make docker-push**
6. Run IMG="your image tag" **make deploy**
7. Apply the yaml manifest from operator/config/samples
8. Check logs of the operator container **kubectl logs -n operator-system operator-controller-manager-xxx**
9. Fix the issue if something goes wrong and repeat.