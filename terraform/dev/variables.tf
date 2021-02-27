variable "project" {
  type = string
}

variable "services" {
  type = list(string)
  default = [
    "cloudfunctions.googleapis.com",
    "cloudbuild.googleapis.com",
  ]
}
