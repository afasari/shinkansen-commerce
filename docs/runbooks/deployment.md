# Deployment Runbook

## Overview

This runbook covers deploying Shinkansen Commerce to production using Kubernetes.

## Pre-Deployment Checklist

- All tests passing
- Code reviewed and approved
- Migration scripts prepared
- Environment variables configured
- Database backups enabled
- Monitoring configured
- Rollback plan documented

## Deployment Steps

### 1. Build Docker Images

```bash
make docker-build
```

### 2. Push to Registry

```bash
make docker-push
```

### 3. Apply Kubernetes Manifests

```bash
cd deploy/k8s
kubectl apply -f namespace.yaml
kubectl apply -f configmaps.yaml
kubectl apply -f secrets.yaml
kubectl apply -f services/
kubectl apply -f deployments/
```

### 4. Verify Deployment

```bash
kubectl get pods -n shinkansen
kubectl logs -f -n shinkansen -l app=gateway
```

## Rolling Updates

### Update Single Service

```bash
kubectl set image deployment/gateway \
  gateway=docker.io/afasari/gateway:v1.0.1 \
  -n shinkansen
```

### Rollback if Issues

```bash
kubectl rollout undo deployment/gateway -n shinkansen
```

## Monitoring After Deployment

### Check Health

```bash
curl https://api.shinkansen.com/health
```

### Check Logs

```bash
kubectl logs -f -n shinkansen --all-containers=true
```

## Troubleshooting

### Pods Not Starting

```bash
kubectl describe pod <pod-name> -n shinkansen
```

### Services Not Accessible

```bash
kubectl get svc -n shinkansen
kubectl get ingress -n shinkansen
```

## Rollback Procedure

### Immediate Rollback

```bash
kubectl rollout undo deployment/gateway -n shinkansen
```

### Full Rollback

```bash
kubectl delete -f deploy/k8s/
kubectl apply -f deploy/k8s/previous/
```
