output "deploy_id" {
  value = "${layer0_deploy.mod.id}"
}

output "load_balancer_id" {
  value = "${layer0_load_balancer.mod.id}"
}

output "load_balancer_url" {
  value = "${layer0_load_balancer.mod.url}"
}

output "service_id" {
  value = "${layer0_service.mod.id}"
}
