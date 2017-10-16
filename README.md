# pile

```
sudo s2i build https://github.com/<username>/pile.git -r master docker.io/<username>/blueprint-go-19-centos7 pile
docker tag pile docker.io/<username>/pile
docker push docker.io/<username>/pile

oc new-app <username>/pile
oc tag --scheduled --source=docker <username>/pile:latest pile:latest
```
or
```
oc new-app <username>/blueprint-go-18-centos7~https://github.com/<username>/pile.git
```
and
```
oc expose service pile
```
