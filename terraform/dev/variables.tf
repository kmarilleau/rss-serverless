variable "project" {
  type = string
}

variable "backend_bucket" {
  type = string
}

variable "region" {
  default = "us-central1"
}

variable "zone" {
  default = "us-central1-c"
}

variable "project_dir" {
  type    = string
  default = "../../"
}
