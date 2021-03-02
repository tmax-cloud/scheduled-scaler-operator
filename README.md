## Overview
The Scheduled-Scaler operator is a controller for CRD ScheduledScaler. With ScheduledScaler resource, you can add Pod scaling schedules in k8s.

## Install

## Usage
1. Sample YAML:
   ```yaml
   apiVersion: tmax.io/v1
    kind: ScheduledScaler
    metadata:
    name: scheduledscaler-sample
    spec:
      target:
        name: test-deployment
      schedule:
        - type: fixed
          runat: '10 * * * * *'
          replicas: 3
        - type: range
          runat: '10 * * * * *'
          minReplicas: 1
          maxReplicas: 3
   ```
   ScheduledScaler is namespaced scope resource. You should define scaling target `Deployment` by specifying spec.target. You should specify spec.schedule to define scaling specification.
   There are two types of scailing; `fixed`, `range`. schedule.runat is when the scaling starts. You need to write it in cron format.

2. Fixed scaling
   `fixed` type of scaling just adjust spec.replicas of the target `Deployment`. In this reason, you need to specify fixed number of `replicas`

3. Range scaling
   `range` type of scaling creates HPA for the target `Deployment`. Thus, you need to specify the range of replicas from `minReplicas` to `maxReplicas`

4. spec.schedule is the list of scaling schedule. These schedules run independently