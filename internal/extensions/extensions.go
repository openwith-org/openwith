package extensions

type Category string

const (
	Config  Category = "Config"
	Docs    Category = "Docs"
	Scripts Category = "Scripts"
	Web     Category = "Web"
	Mobile  Category = "Mobile"
	Systems Category = "Systems"
	DB      Category = "DB"
	DevOps  Category = "DevOps"
	Data    Category = "Data"
	Images  Category = "Images"
	Video   Category = "Video"
	Audio   Category = "Audio"
	PDFs    Category = "PDFs"
	Other   Category = "Other"
)

type Extension struct {
	Ext      string
	Category Category
}

type Group struct {
	Category   Category
	Extensions []Extension
}

var CategoryOrder = []Category{Config, Docs, Scripts, Web, Mobile, Systems, DB, Data, DevOps, Images, Video, Audio, PDFs, Other}

var All = []Extension{
	// Config
	{".json", Config}, {".jsonc", Config}, {".yaml", Config}, {".yml", Config},
	{".toml", Config}, {".ini", Config}, {".cfg", Config}, {".env", Config},
	{".nix", Config},
	// Docs
	{".md", Docs}, {".txt", Docs}, {".log", Docs},
	// Scripts
	{".sh", Scripts}, {".bash", Scripts}, {".zsh", Scripts}, {".fish", Scripts},
	{".py", Scripts}, {".rb", Scripts}, {".rs", Scripts}, {".go", Scripts},
	{".lua", Scripts}, {".zig", Scripts}, {".nim", Scripts},
	// Web
	{".js", Web}, {".jsx", Web}, {".ts", Web}, {".tsx", Web},
	{".html", Web}, {".css", Web}, {".scss", Web},
	{".vue", Web}, {".svelte", Web}, {".astro", Web},
	// Mobile
	{".dart", Mobile}, {".swift", Mobile}, {".kt", Mobile}, {".java", Mobile},
	// Systems
	{".c", Systems}, {".cpp", Systems}, {".h", Systems}, {".hpp", Systems},
	{".proto", Systems},
	// DB
	{".sql", DB}, {".graphql", DB}, {".prisma", DB},
	// Data
	{".csv", Data}, {".xml", Data}, {".svg", Data},
	// DevOps
	{".dockerfile", DevOps}, {".tf", DevOps}, {".hcl", DevOps},
	// Other
	{".lock", Other}, {".gitignore", Other}, {".editorconfig", Other},
}

// MediaAll contains media/document file extensions (non-code).
var MediaAll = []Extension{
	// Images
	{".png", Images}, {".jpg", Images}, {".jpeg", Images}, {".gif", Images},
	{".webp", Images}, {".tiff", Images}, {".bmp", Images}, {".ico", Images},
	{".heic", Images}, {".psd", Images},
	// Video
	{".mp4", Video}, {".mkv", Video}, {".avi", Video}, {".mov", Video},
	{".wmv", Video}, {".webm", Video}, {".flv", Video},
	// Audio
	{".mp3", Audio}, {".flac", Audio}, {".wav", Audio}, {".aac", Audio},
	{".ogg", Audio}, {".m4a", Audio}, {".wma", Audio},
	// PDFs
	{".pdf", PDFs},
}

// EnableMediaExtensions adds media file extensions to the main list.
func EnableMediaExtensions() {
	Merge(MediaAll)
}

// Merge appends extra extensions, skipping duplicates.
func Merge(extras []Extension) {
	existing := make(map[string]bool)
	for _, e := range All {
		existing[e.Ext] = true
	}
	for _, e := range extras {
		if !existing[e.Ext] {
			All = append(All, e)
			existing[e.Ext] = true
		}
	}
}

func AllExts() []string {
	exts := make([]string, len(All))
	for i, e := range All {
		exts[i] = e.Ext
	}
	return exts
}

func Grouped() []Group {
	byCategory := make(map[Category][]Extension)
	for _, ext := range All {
		byCategory[ext.Category] = append(byCategory[ext.Category], ext)
	}
	groups := make([]Group, 0, len(CategoryOrder))
	for _, cat := range CategoryOrder {
		if exts, ok := byCategory[cat]; ok {
			groups = append(groups, Group{Category: cat, Extensions: exts})
		}
	}
	return groups
}
