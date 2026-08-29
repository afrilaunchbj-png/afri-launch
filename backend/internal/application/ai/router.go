// Package ai orchestre l'accès aux providers IA (OpenAI + HeyGen) et le
// routage de modèle par tâche. Voir docs/ai.md.
package ai

// Task identifie un type de génération, pour choisir le bon modèle.
type Task string

const (
	TaskResearch Task = "research" // recherche de marché (contenu long)
	TaskContent  Task = "content"  // contenu long (ebook, chapitres)
	TaskIdeation Task = "ideation" // génération d'idées (léger/rapide)
	TaskImage    Task = "image"    // génération d'image
)

// ModelRouter mappe une tâche vers un nom de modèle.
type ModelRouter struct {
	research string
	ideation string
	image    string
}

// NewModelRouter construit le routeur depuis la config (noms de modèles).
func NewModelRouter(research, ideation, image string) ModelRouter {
	return ModelRouter{research: research, ideation: ideation, image: image}
}

// ModelFor renvoie le modèle adapté à la tâche.
func (r ModelRouter) ModelFor(t Task) string {
	switch t {
	case TaskIdeation:
		return r.ideation
	case TaskImage:
		return r.image
	default: // research, content, et tout le reste → modèle "long content"
		return r.research
	}
}
