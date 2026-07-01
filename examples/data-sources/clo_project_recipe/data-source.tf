# Resolve a recipe_id by name to pass to clo_compute_instance.
data "clo_project_recipe" "web" {
  project_id = "1e8ff0f7-0b8c-4ec5-a0a4-e30cea0db287"
  name       = "wordpress"
}