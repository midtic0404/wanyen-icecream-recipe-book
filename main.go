package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Recipe struct {
	ID           int       `json:"id"`
	Name         string    `json:"name"`
	Ingredients  string    `json:"ingredients"`
	Instructions string    `json:"instructions"`
	PrepTime     int       `json:"prep_time"`
	CreatedAt    time.Time `json:"created_at"`
}

var db *sql.DB

func initDB() {
	var err error
	db, err = sql.Open("sqlite3", "./recipes.db")
	if err != nil {
		log.Fatal(err)
	}

	createTable := `
	CREATE TABLE IF NOT EXISTS recipes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		ingredients TEXT NOT NULL,
		instructions TEXT NOT NULL,
		prep_time INTEGER,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	_, err = db.Exec(createTable)
	if err != nil {
		log.Fatal(err)
	}

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM recipes").Scan(&count)
	if err != nil {
		log.Fatal(err)
	}

	if count == 0 {
		insertSampleData()
	}
}

func insertSampleData() {
	sampleRecipes := []Recipe{
		{
			Name:         "Vanilla Bean",
			Ingredients:  "2 cups heavy cream\n1 cup milk\n3/4 cup sugar\n1 vanilla bean\n6 egg yolks",
			Instructions: "Heat cream and milk\nWhisk egg yolks with sugar\nTemper eggs with hot cream\nCook until thick\nStrain and chill\nChurn in ice cream maker",
			PrepTime:     45,
		},
		{
			Name:         "Chocolate Fudge",
			Ingredients:  "2 cups heavy cream\n1 cup milk\n3/4 cup sugar\n1/2 cup cocoa powder\n6 egg yolks\n1/2 cup fudge sauce",
			Instructions: "Whisk cocoa with sugar\nHeat cream and milk\nWhisk egg yolks\nTemper eggs with hot cream mixture\nCook until thick\nAdd fudge swirls\nChurn in ice cream maker",
			PrepTime:     50,
		},
	}

	for _, recipe := range sampleRecipes {
		createRecipe(recipe)
	}
}

func getAllRecipes() ([]Recipe, error) {
	rows, err := db.Query("SELECT id, name, ingredients, instructions, prep_time, created_at FROM recipes ORDER BY created_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var recipes []Recipe
	for rows.Next() {
		var recipe Recipe
		err := rows.Scan(&recipe.ID, &recipe.Name, &recipe.Ingredients, &recipe.Instructions, &recipe.PrepTime, &recipe.CreatedAt)
		if err != nil {
			return nil, err
		}
		recipes = append(recipes, recipe)
	}
	return recipes, nil
}

func createRecipe(recipe Recipe) error {
	_, err := db.Exec(
		"INSERT INTO recipes (name, ingredients, instructions, prep_time) VALUES (?, ?, ?, ?)",
		recipe.Name, recipe.Ingredients, recipe.Instructions, recipe.PrepTime,
	)
	return err
}

func getRecipeByID(id int) (*Recipe, error) {
	var recipe Recipe
	err := db.QueryRow(
		"SELECT id, name, ingredients, instructions, prep_time, created_at FROM recipes WHERE id = ?",
		id,
	).Scan(&recipe.ID, &recipe.Name, &recipe.Ingredients, &recipe.Instructions, &recipe.PrepTime, &recipe.CreatedAt)
	
	if err != nil {
		return nil, err
	}
	return &recipe, nil
}

func updateRecipe(recipe Recipe) error {
	_, err := db.Exec(
		"UPDATE recipes SET name = ?, ingredients = ?, instructions = ?, prep_time = ? WHERE id = ?",
		recipe.Name, recipe.Ingredients, recipe.Instructions, recipe.PrepTime, recipe.ID,
	)
	return err
}

func deleteRecipe(id int) error {
	_, err := db.Exec("DELETE FROM recipes WHERE id = ?", id)
	return err
}

func main() {
	initDB()
	defer db.Close()

	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/static/", staticHandler)
	http.HandleFunc("/recipe/", recipeDetailHandler)
	http.HandleFunc("/add-recipe", addRecipeFormHandler)
	http.HandleFunc("/edit-recipe/", editRecipeHandler)
	http.HandleFunc("/delete-recipe/", deleteRecipeHandler)
	
	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	funcMap := template.FuncMap{
		"split": strings.Split,
	}
	tmpl := template.Must(template.New("").Funcs(funcMap).ParseFiles("templates/base.html", "templates/home.html"))
	
	recipes, err := getAllRecipes()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	data := struct {
		Title   string
		Recipes []Recipe
	}{
		Title:   "Ice Cream Recipe Book",
		Recipes: recipes,
	}
	
	if err := tmpl.ExecuteTemplate(w, "base.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func staticHandler(w http.ResponseWriter, r *http.Request) {
	http.StripPrefix("/static/", http.FileServer(http.Dir("static/"))).ServeHTTP(w, r)
}

func recipeDetailHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	
	idStr := strings.TrimPrefix(r.URL.Path, "/recipe/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid recipe ID", http.StatusBadRequest)
		return
	}
	
	recipe, err := getRecipeByID(id)
	if err != nil {
		http.Error(w, "Recipe not found", http.StatusNotFound)
		return
	}
	
	ingredientsList := strings.Split(recipe.Ingredients, "\n")
	instructionsList := strings.Split(recipe.Instructions, "\n")
	
	html := `
	<div class="modal-content">
		<button onclick="this.closest('.modal').style.display='none'" style="float: right; background: none; border: none; font-size: 1.5rem; cursor: pointer;">&times;</button>
		<h3>` + recipe.Name + `</h3>
		<p><strong>Prep Time:</strong> ` + strconv.Itoa(recipe.PrepTime) + ` minutes</p>
		<p><strong>Created:</strong> ` + recipe.CreatedAt.Format("January 2, 2006") + `</p>
		
		<div style="margin: 1rem 0;">
			<h4>Ingredients:</h4>
			<ul>`
	
	for _, ingredient := range ingredientsList {
		if strings.TrimSpace(ingredient) != "" {
			html += `<li>` + strings.TrimSpace(ingredient) + `</li>`
		}
	}
	
	html += `</ul>
		</div>
		
		<div style="margin: 1rem 0;">
			<h4>Instructions:</h4>
			<ol>`
	
	for _, instruction := range instructionsList {
		if strings.TrimSpace(instruction) != "" {
			html += `<li>` + strings.TrimSpace(instruction) + `</li>`
		}
	}
	
	html += `</ol>
		</div>
		
		<div style="display: flex; gap: 1rem; margin-top: 2rem;">
			<button hx-get="/edit-recipe/` + strconv.Itoa(recipe.ID) + `" hx-target="#recipe-detail" class="btn btn-primary">Edit Recipe</button>
			<button hx-delete="/delete-recipe/` + strconv.Itoa(recipe.ID) + `" hx-target="#recipe-list" hx-swap="outerHTML" hx-confirm="Are you sure you want to delete this recipe?" class="btn btn-secondary" style="background: #dc3545; color: white; border-color: #dc3545;">Delete Recipe</button>
			<button onclick="this.closest('.modal').style.display='none'" class="btn btn-secondary">Close</button>
		</div>
	</div>
	<script>
		document.querySelector('#recipe-detail').style.display = 'flex';
	</script>`
	
	w.Write([]byte(html))
}

func editRecipeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	
	idStr := strings.TrimPrefix(r.URL.Path, "/edit-recipe/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid recipe ID", http.StatusBadRequest)
		return
	}
	
	if r.Method == "GET" {
		recipe, err := getRecipeByID(id)
		if err != nil {
			http.Error(w, "Recipe not found", http.StatusNotFound)
			return
		}
		
		html := `
		<div class="modal-content">
			<button onclick="this.closest('.modal').style.display='none'" style="float: right; background: none; border: none; font-size: 1.5rem; cursor: pointer;">&times;</button>
			<h3>Edit Recipe</h3>
			<form hx-put="/edit-recipe/` + strconv.Itoa(recipe.ID) + `" hx-target="#recipe-list" hx-swap="outerHTML">
				<div style="margin-bottom: 1rem;">
					<label for="name" style="display: block; margin-bottom: 0.5rem; font-weight: 600;">Recipe Name:</label>
					<input type="text" id="name" name="name" value="` + recipe.Name + `" required style="width: 100%; padding: 8px; border: 1px solid #ddd; border-radius: 4px;">
				</div>
				<div style="margin-bottom: 1rem;">
					<label for="ingredients" style="display: block; margin-bottom: 0.5rem; font-weight: 600;">Ingredients (one per line):</label>
					<textarea id="ingredients" name="ingredients" required style="width: 100%; padding: 8px; border: 1px solid #ddd; border-radius: 4px; height: 100px;">` + recipe.Ingredients + `</textarea>
				</div>
				<div style="margin-bottom: 1rem;">
					<label for="instructions" style="display: block; margin-bottom: 0.5rem; font-weight: 600;">Instructions (one per line):</label>
					<textarea id="instructions" name="instructions" required style="width: 100%; padding: 8px; border: 1px solid #ddd; border-radius: 4px; height: 100px;">` + recipe.Instructions + `</textarea>
				</div>
				<div style="margin-bottom: 1rem;">
					<label for="prep_time" style="display: block; margin-bottom: 0.5rem; font-weight: 600;">Prep Time (minutes):</label>
					<input type="number" id="prep_time" name="prep_time" value="` + strconv.Itoa(recipe.PrepTime) + `" min="1" style="width: 100%; padding: 8px; border: 1px solid #ddd; border-radius: 4px;">
				</div>
				<div style="display: flex; gap: 1rem;">
					<button type="submit" class="btn btn-primary">Update Recipe</button>
					<button type="button" onclick="this.closest('.modal').style.display='none'" class="btn btn-secondary">Cancel</button>
				</div>
			</form>
		</div>
		<script>
			document.querySelector('#recipe-detail').style.display = 'flex';
		</script>`
		
		w.Write([]byte(html))
	} else if r.Method == "PUT" {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		
		prepTime, _ := strconv.Atoi(r.FormValue("prep_time"))
		
		updatedRecipe := Recipe{
			ID:           id,
			Name:         r.FormValue("name"),
			Ingredients:  r.FormValue("ingredients"),
			Instructions: r.FormValue("instructions"),
			PrepTime:     prepTime,
		}
		
		err = updateRecipe(updatedRecipe)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		
		recipes, err := getAllRecipes()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		
		html := `<div id="recipe-list" class="recipe-grid">`
		for _, recipe := range recipes {
			ingredientsList := strings.Split(recipe.Ingredients, "\n")
			html += `
			<div class="recipe-card">
				<h3>` + recipe.Name + `</h3>
				<p><strong>Prep Time:</strong> ` + strconv.Itoa(recipe.PrepTime) + ` minutes</p>
				<p><strong>Ingredients:</strong></p>
				<ul>`
			for _, ingredient := range ingredientsList {
				if strings.TrimSpace(ingredient) != "" {
					html += `<li>` + strings.TrimSpace(ingredient) + `</li>`
				}
			}
			html += `</ul>
				<button hx-get="/recipe/` + strconv.Itoa(recipe.ID) + `" hx-target="#recipe-detail" class="btn btn-primary">View Details</button>
			</div>`
		}
		html += `</div>
		<script>
			document.querySelector('#recipe-detail').style.display = 'none';
		</script>`
		
		w.Write([]byte(html))
	}
}

func deleteRecipeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	
	if r.Method != "DELETE" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	idStr := strings.TrimPrefix(r.URL.Path, "/delete-recipe/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid recipe ID", http.StatusBadRequest)
		return
	}
	
	err = deleteRecipe(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	recipes, err := getAllRecipes()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	html := `<div id="recipe-list" class="recipe-grid">`
	for _, recipe := range recipes {
		ingredientsList := strings.Split(recipe.Ingredients, "\n")
		html += `
		<div class="recipe-card">
			<h3>` + recipe.Name + `</h3>
			<p><strong>Prep Time:</strong> ` + strconv.Itoa(recipe.PrepTime) + ` minutes</p>
			<p><strong>Ingredients:</strong></p>
			<ul>`
		for _, ingredient := range ingredientsList {
			if strings.TrimSpace(ingredient) != "" {
				html += `<li>` + strings.TrimSpace(ingredient) + `</li>`
			}
		}
		html += `</ul>
			<button hx-get="/recipe/` + strconv.Itoa(recipe.ID) + `" hx-target="#recipe-detail" class="btn btn-primary">View Details</button>
		</div>`
	}
	html += `</div>
	<script>
		document.querySelector('#recipe-detail').style.display = 'none';
	</script>`
	
	w.Write([]byte(html))
}

func addRecipeFormHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	
	if r.Method == "GET" {
		html := `
		<div class="modal-content">
			<button onclick="this.closest('.modal').style.display='none'" style="float: right; background: none; border: none; font-size: 1.5rem; cursor: pointer;">&times;</button>
			<h3>Add New Recipe</h3>
			<form hx-post="/add-recipe" hx-target="#recipe-list" hx-swap="outerHTML">
				<div style="margin-bottom: 1rem;">
					<label for="name" style="display: block; margin-bottom: 0.5rem; font-weight: 600;">Recipe Name:</label>
					<input type="text" id="name" name="name" required style="width: 100%; padding: 8px; border: 1px solid #ddd; border-radius: 4px;">
				</div>
				<div style="margin-bottom: 1rem;">
					<label for="ingredients" style="display: block; margin-bottom: 0.5rem; font-weight: 600;">Ingredients (one per line):</label>
					<textarea id="ingredients" name="ingredients" required style="width: 100%; padding: 8px; border: 1px solid #ddd; border-radius: 4px; height: 100px;" placeholder="2 cups heavy cream&#10;1 cup milk&#10;3/4 cup sugar"></textarea>
				</div>
				<div style="margin-bottom: 1rem;">
					<label for="instructions" style="display: block; margin-bottom: 0.5rem; font-weight: 600;">Instructions (one per line):</label>
					<textarea id="instructions" name="instructions" required style="width: 100%; padding: 8px; border: 1px solid #ddd; border-radius: 4px; height: 100px;" placeholder="Heat cream and milk&#10;Whisk egg yolks with sugar&#10;Cook until thick"></textarea>
				</div>
				<div style="margin-bottom: 1rem;">
					<label for="prep_time" style="display: block; margin-bottom: 0.5rem; font-weight: 600;">Prep Time (minutes):</label>
					<input type="number" id="prep_time" name="prep_time" min="1" style="width: 100%; padding: 8px; border: 1px solid #ddd; border-radius: 4px;">
				</div>
				<div style="display: flex; gap: 1rem;">
					<button type="submit" class="btn btn-primary">Add Recipe</button>
					<button type="button" onclick="this.closest('.modal').style.display='none'" class="btn btn-secondary">Cancel</button>
				</div>
			</form>
		</div>
		<script>
			document.querySelector('#recipe-modal').style.display = 'flex';
		</script>`
		
		w.Write([]byte(html))
	} else if r.Method == "POST" {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		
		prepTime, _ := strconv.Atoi(r.FormValue("prep_time"))
		
		newRecipe := Recipe{
			Name:         r.FormValue("name"),
			Ingredients:  r.FormValue("ingredients"),
			Instructions: r.FormValue("instructions"),
			PrepTime:     prepTime,
		}
		
		err = createRecipe(newRecipe)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		
		recipes, err := getAllRecipes()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		
		html := `<div id="recipe-list" class="recipe-grid">`
		for _, recipe := range recipes {
			ingredientsList := strings.Split(recipe.Ingredients, "\n")
			html += `
			<div class="recipe-card">
				<h3>` + recipe.Name + `</h3>
				<p><strong>Prep Time:</strong> ` + strconv.Itoa(recipe.PrepTime) + ` minutes</p>
				<p><strong>Ingredients:</strong></p>
				<ul>`
			for _, ingredient := range ingredientsList {
				if strings.TrimSpace(ingredient) != "" {
					html += `<li>` + strings.TrimSpace(ingredient) + `</li>`
				}
			}
			html += `</ul>
				<button hx-get="/recipe/` + strconv.Itoa(recipe.ID) + `" hx-target="#recipe-detail" class="btn btn-primary">View Details</button>
			</div>`
		}
		html += `</div>
		<script>
			document.querySelector('#recipe-modal').style.display = 'none';
		</script>`
		
		w.Write([]byte(html))
	}
}