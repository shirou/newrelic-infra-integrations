============================
New Relic Infra Integrations
============================

This repository includes unofficial New Relic Infra Integrations.

Available Integrations
--------------------------

check_tcp
  check tcp connection is available

How to use
-------------

1. go build

2. place exec binary and *-definition.yml to `/var/db/newrelic-infra/custom-integrations`

3. edit *-config.yml and place /etc/newrelic-infra/integrations.d.

4. restart newrelic-infra agent


LICENSE
-----------

Apache License 2.0
