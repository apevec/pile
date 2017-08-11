# pile

```
sudo s2i build https://github.com/arapov/pile.git -r master docker.io/arapov/blueprint-go-18-centos7 pile
docker tag pile docker.io/arapov/pile
docker push docker.io/arapov/pile

oc new-app arapov/pile
oc tag --scheduled --source=docker arapov/pile:latest pile:latest
```
or
```
oc new-app arapov/blueprint-go-18-centos7~https://github.com/arapov/pile.git
```
and
```
oc expose service pile
```
