module "feeds_fetcher" {
  source  = "./modules/function"
  project = var.project

  function_source      = "${var.project_dir}feeds-fetcher/fetcher"
  function_name        = "feeds-fetcher"
  function_runtime     = "go113"
  function_entry_point = "FetchURLAndStoreItsContent"
}
