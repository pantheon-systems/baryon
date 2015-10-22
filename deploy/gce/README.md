# baryon controller
this is the kube  Replication Controller config. The RC can be recreated from this file if need-be using
```
  kubectl create -f baryon-rc.yml
```
There is also  Service that associates to the baryon pod/rc:

```
  kubectl create -f baryon-service.yml
```

If they already exist you can delete either of them first with `kubectl delete rc baryon` or `kubectl delete service baryon`

## secrets
There is a secrets volume attached to baryon for getting its certs for ssl. This secrets.yml is in a onelogin note named `baryon-ssl.yml` also found here: https://pantheon.onelogin.com/notes/18388/

After placing this in your directory you would want to delete the old secrete, add  new, and redeploy baryon:

Find the secret
```
─➤ kubectl get secret
NAME                  TYPE                                  DATA
baryon-ssl            Opaque                                2
default-token-pussp   kubernetes.io/service-account-token   2
```

Delete the existing secret
```
─➤ kubectl delete secret baryon-ssl
secrets/baryon-ssl
```

Create the new secret
```
─➤  kubectl create -f baryon-ssl.yml
secrets/baryon-ssl
```

Get the pod id
```
─➤  kubectl get pod
NAME           READY     STATUS    RESTARTS   AGE
baryon-n5nli   1/1       Running   2          1h
```


Relaunch the pod
```
─➤  kubectl delete pod baryon-n5nli
pods/baryon-n5nli
```

All done!

you can verify the pod has relaunched
```
─➤  kubectl get pod
NAME           READY     STATUS    RESTARTS   AGE
baryon-ua7xo   0/1       Running   0          5s
```
