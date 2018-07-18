module "machinebox" {
  source = "/Users/nicj/Developer/terraform/terraform-modules/elasticbeanstalk-docker"

  instance_type = "t2.micro"

  application_name        = "machinebox"
  application_description = "Machinebox server"
  application_environment = "development"
  application_version     = "1.0.0"
  docker_image            = "machinebox/facebox"
  docker_tag              = "latest"
  docker_ports            = ["8080"]
  health_check            = "/info"
  env_vars                = ["MB_KEY", "${var.mb_key}"]
  elb_scheme              = "external"
}

module "emojify" {
  source = "/Users/nicj/Developer/terraform/terraform-modules/elasticbeanstalk-docker"

  instance_type = "t2.nano"

  application_name        = "emojify"
  application_description = "Emojify server"
  application_environment = "development"
  application_version     = "1.0.0"
  docker_image            = "docker.io/nicholasjackson/emojify"
  docker_tag              = "latest"
  docker_ports            = ["9090"]
  health_check            = "/health"
  env_vars                = ["FACEBOX", "http://${module.machinebox.cname}"]
  elb_scheme              = "external"
}

output "elasticbeanstalk_emojify_cname" {
  value = "${module.emojify.cname}"
}

output "elasticbeanstalk_machinebox_cname" {
  value = "${module.machinebox.cname}"
}
