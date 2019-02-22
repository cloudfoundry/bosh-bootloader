#!/bin/sh

generate_config() {
  local index=$1

  cat > /tmp/config$index.yml <<EOF
vm_extensions:
- name: cf-router-network-properties-${index}
  cloud_properties:
    backend_service: ((router_backend_service))
    target_pool: ((ws_target_pool))
    tags:
    - ((router_backend_service))
    - ((ws_target_pool))

- name: diego-ssh-proxy-network-properties-${index}
  cloud_properties:
    target_pool: ((ssh_proxy_target_pool))
    tags:
    - ((ssh_proxy_target_pool))

- name: cf-tcp-router-network-properties-${index}
  cloud_properties:
    target_pool: ((tcp_router_target_pool))
    tags:
    - ((tcp_router_target_pool))
EOF
}

vars() {
  bosh int vars/cloud-config-vars.yml --path /$1/$2
}

main() {
  local cf_env_count=${TF_VAR_cf_env_count}

  for i in `seq 0 $((cf_env_count-1))`; do
    generate_config $i

    bosh update-config -n \
      --type cloud \
      --name "cf$i" \
      /tmp/config$i.yml \
      -v router_backend_service=$(vars router_backend_services $i) \
      -v ws_target_pool=$(vars ws_target_pools $i) \
      -v ssh_proxy_target_pool=$(vars ssh_proxy_target_pools $i) \
      -v tcp_router_target_pool=$(vars tcp_router_target_pools $i)
  done
}

main "$@"
