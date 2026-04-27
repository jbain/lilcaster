package ffmpeg

import "math/rand"

var phraseColors = []string{
	"white", "yellow", "cyan", "lime", "orange", "magenta",
	"red", "lightblue", "lightgreen", "gold", "coral", "violet",
	"pink", "turquoise", "springgreen", "aquamarine", "chartreuse",
	"hotpink", "deepskyblue", "lightskyblue",
}

// wordList is a ~1000-word safe, work- and child-friendly vocabulary for random phrase generation.
var wordList = []string{
	// Animals
	"ant", "ape", "armadillo", "badger", "bat", "bear", "beaver", "bee", "beetle", "bison",
	"boar", "butterfly", "calf", "camel", "carp", "cat", "cheetah", "chimp", "chipmunk",
	"clam", "cobra", "colt", "crab", "crane", "cricket", "crow", "deer", "dog", "dolphin",
	"dove", "dragonfly", "duck", "eagle", "eel", "elk", "emu", "falcon", "ferret", "finch",
	"firefly", "fox", "frog", "gecko", "giraffe", "goat", "goose", "gorilla", "grasshopper",
	"hare", "hawk", "hedgehog", "heron", "ibis", "iguana", "jaguar", "jay", "koala", "lamb",
	"lark", "lemur", "lion", "llama", "lobster", "lynx", "mink", "mole", "moose", "moth",
	"mouse", "newt", "octopus", "otter", "owl", "oyster", "panda", "parrot", "penguin",
	"perch", "pike", "porcupine", "puma", "rabbit", "raccoon", "ram", "raven", "robin",
	"seal", "shark", "shrew", "skunk", "sloth", "snail", "sparrow", "spider", "squid",
	"squirrel", "starfish", "stork", "swan", "tiger", "toad", "trout", "turtle", "vole",
	"vulture", "walrus", "wasp", "weasel", "whale", "wolf", "wolverine", "wren", "yak", "zebra",
	"albatross", "bluebird", "bullfrog", "condor", "cormorant", "coyote", "crocodile",
	"cuckoo", "egret", "flamingo", "gnu", "hamster", "herring", "hummingbird", "impala",
	"kestrel", "kingfisher", "macaw", "magpie", "marmot", "meerkat", "mongoose", "narwhal",
	"nightingale", "ocelot", "parakeet", "peacock", "pelican", "pheasant", "platypus",
	"puffin", "quail", "salamander", "seahorse", "swallow", "tapir", "tern", "thrush",
	"toucan", "warbler", "woodpecker",

	// Plants, Trees, Flowers
	"acorn", "aloe", "ash", "aspen", "bamboo", "birch", "blossom", "bud", "cactus", "cedar",
	"clover", "cone", "daisy", "dandelion", "elm", "fern", "fir", "flax", "frond", "gorse",
	"grove", "hazel", "heath", "herb", "holly", "hyacinth", "iris", "ivy", "jasmine",
	"juniper", "kelp", "laurel", "leaf", "lilac", "lily", "linden", "log", "maple", "mint",
	"moss", "mushroom", "myrtle", "nettle", "oak", "orchid", "palm", "petal", "pine",
	"poplar", "pollen", "reed", "root", "rose", "rowan", "seed", "shrub", "sorrel",
	"sprout", "stem", "thistle", "thorn", "thyme", "tulip", "vine", "willow", "wisteria",
	"aster", "azalea", "basil", "begonia", "bracken", "broom", "catkin", "chicory",
	"cress", "heather", "hops", "marjoram", "primrose", "rosemary", "rue", "tansy",
	"verbena", "yarrow",

	// Weather, Sky, Space
	"aurora", "blizzard", "breeze", "cirrus", "cloud", "comet", "cosmos", "crater", "dawn",
	"dusk", "eclipse", "equinox", "flurry", "fog", "frost", "gale", "galaxy", "gust", "hail",
	"haze", "ice", "lightning", "lunar", "meteor", "mist", "moon", "nebula", "nova", "orbit",
	"prism", "quasar", "rain", "rainbow", "sleet", "snow", "solstice", "star", "storm", "sun",
	"sunrise", "sunset", "thunder", "tornado", "twilight", "vortex", "wind", "zenith",
	"altitude", "atmosphere", "cumulus", "cyclone", "nimbostratus", "stratosphere", "stratus",
	"zephyr",

	// Geography, Terrain
	"basin", "bay", "beach", "bog", "brook", "butte", "canyon", "cape", "cave", "channel",
	"cliff", "coast", "crag", "creek", "delta", "desert", "dune", "estuary", "fjord", "ford",
	"geyser", "glacier", "glen", "gorge", "gulf", "harbor", "headland", "hill", "hollow",
	"inlet", "island", "lagoon", "lake", "ledge", "marsh", "mesa", "moor", "mount", "oasis",
	"ocean", "peak", "plain", "plateau", "pond", "pool", "prairie", "ravine", "reef", "ridge",
	"river", "sea", "shoal", "shore", "slope", "spring", "steppe", "stream", "summit", "tide",
	"trail", "vale", "valley", "volcano", "waterfall", "wetland",
	"atoll", "bayou", "bluff", "cove", "fell", "grotto", "jungle", "loch", "pampa",
	"savanna", "scrub", "tundra", "upland", "weald",

	// Colors
	"amber", "amethyst", "azure", "beige", "cobalt", "coral", "crimson", "cyan", "ebony",
	"fuchsia", "gold", "gray", "indigo", "ivory", "jade", "lavender", "lime", "magenta",
	"maroon", "mauve", "navy", "ochre", "olive", "onyx", "orange", "peach", "pearl",
	"pink", "plum", "purple", "rose", "ruby", "russet", "sapphire", "scarlet", "sienna",
	"silver", "tan", "taupe", "teal", "topaz", "turquoise", "violet", "white",

	// Objects, Tools
	"anchor", "anvil", "arrow", "axle", "barrel", "basket", "beacon", "bell", "blade",
	"bolt", "bowl", "bridge", "candle", "canvas", "cart", "chain", "chisel", "cog",
	"compass", "crown", "dial", "dome", "drum", "flask", "flag", "flute", "funnel", "gear",
	"globe", "gavel", "hammer", "helm", "hilt", "hook", "jar", "kite", "knot", "lantern",
	"latch", "lever", "loom", "mallet", "map", "mast", "medal", "mirror", "needle",
	"notch", "oar", "orb", "paddle", "pail", "pan", "pin", "pipe", "plank", "pulley",
	"quill", "rod", "rope", "rudder", "sail", "spool", "staff", "sundial", "torch",
	"trident", "vault", "wheel", "whistle", "windmill", "buoy", "cradle", "mortar", "prow",

	// Adjectives
	"active", "agile", "alert", "ancient", "ardent", "bold", "brave", "bright", "brisk",
	"calm", "careful", "clear", "clever", "crisp", "daring", "deep", "dense", "distant",
	"durable", "earnest", "fair", "fast", "fierce", "fine", "firm", "fleet", "fresh",
	"gentle", "glad", "grand", "golden", "happy", "hardy", "hearty", "jolly", "jovial",
	"just", "keen", "kind", "large", "lasting", "lively", "lofty", "luminous", "lucky",
	"mighty", "mild", "misty", "noble", "nimble", "open", "patient", "peaceful", "polished",
	"prompt", "quick", "quiet", "radiant", "rapid", "rare", "ready", "regal", "robust",
	"safe", "serene", "sharp", "silent", "simple", "sleek", "slim", "smooth", "solid",
	"spirited", "steady", "still", "strong", "sturdy", "swift", "tall", "tidy", "tiny",
	"tranquil", "true", "vast", "warm", "wide", "wise", "young", "zealous",
	"able", "airy", "breezy", "buoyant", "candid", "capable", "carefree", "cheerful",
	"chipper", "civil", "cozy", "dear", "dignified", "eager", "fearless", "fond",
	"graceful", "hale", "ideal", "jaunty", "kindly", "lean", "mellow", "natural", "neat",
	"orderly", "plain", "prim", "proper", "prudent", "pure", "quaint", "refined",
	"reliable", "restful", "rosy", "shiny", "silky", "sincere", "snappy", "sterling",
	"sunny", "sweet", "thorough", "trusty", "upbeat", "wholesome", "zesty",

	// Actions, Movement
	"bloom", "bound", "circle", "climb", "coast", "craft", "creep", "cruise", "dance",
	"dash", "dive", "drift", "float", "flow", "flutter", "fly", "glide", "glow", "grow",
	"hike", "jump", "leap", "march", "plunge", "race", "ride", "roam", "rush", "shimmer",
	"shine", "skip", "soar", "sprint", "stride", "trek", "wade", "wander",
	"amble", "ascend", "burst", "cascade", "cavort", "dawdle", "frolic", "gallop",
	"gambol", "glisten", "hasten", "hover", "prance", "prowl", "ramble", "saunter",
	"scramble", "slink", "stroll", "swoop", "tumble", "whirl", "zoom",

	// Food, Flavors, Botanicals
	"almond", "apricot", "barley", "cashew", "cherry", "chestnut", "chive", "clove",
	"coconut", "dill", "fennel", "fig", "ginger", "grape", "hawthorn", "honey",
	"huckleberry", "kale", "lemon", "lentil", "mango", "melon", "nutmeg", "oregano",
	"papaya", "parsley", "peach", "pear", "pepper", "quince", "radish", "raisin",
	"sage", "sesame", "tarragon", "vanilla", "walnut", "wheat",
	"anise", "blueberry", "cardamom", "caraway", "carob", "citrus", "cranberry",
	"currant", "elderberry", "gooseberry", "kumquat", "licorice", "mulberry", "paprika",
	"pistachio", "saffron", "strawberry", "turmeric",

	// Buildings, Structures
	"arch", "atrium", "battlement", "castle", "chapel", "citadel", "cloister",
	"fortress", "fountain", "gateway", "manor", "moat", "pagoda", "pavilion",
	"rampart", "rotunda", "spire", "steeple", "tower", "turret",
	"lodge", "lookout", "outpost", "porch", "sanctuary", "shelter", "shrine", "stockade",

	// Gems, Minerals
	"agate", "basalt", "beryl", "calcite", "crystal", "diamond", "flint", "garnet",
	"jasper", "lapis", "marble", "obsidian", "opal", "pyrite", "quartz", "schist",
	"shale", "zircon",

	// Music
	"aria", "ballad", "chord", "coda", "fugue", "harmony", "lute", "melody", "motif",
	"nocturne", "opus", "refrain", "sonata", "verse",

	// Celestial, Navigation
	"apogee", "axis", "corona", "crest", "flare", "helios", "horizon", "meridian",
	"nadir", "perigee", "polar",
	"cardinal", "eastern", "northern", "southern", "western", "azimuth", "heading",
	"latitude", "outbound", "pathway", "waymark",

	// Textures, Qualities
	"crispy", "downy", "dusty", "earthy", "feathery", "fleecy", "fluffy", "foamy",
	"furry", "gauzy", "grainy", "gritty", "leafy", "leathery", "lithe", "lumpy",
	"mossy", "muddy", "pearly", "powdery", "puffy", "rocky", "satiny", "scaly",
	"spongy", "stony", "velvety", "wispy", "woody", "woolly",

	// Time, Seasons
	"almanac", "calendar", "chronicle", "cycle", "decade", "epoch", "era",
	"evening", "morning", "seasonal", "summer", "weekly", "yearly",

	// Poetic, Misc
	"abyss", "acclaim", "adrift", "aloft", "astir", "aura", "cipher", "clarity",
	"codex", "daylight", "denizen", "elegy", "embers", "encore", "essence", "eternal",
	"ethos", "haven", "hinterland", "inspire", "light", "lore", "mystique", "nature",
	"nightfall", "onward", "passage", "path", "radiance", "realm", "rhythm", "rise",
	"roving", "sable", "shadow", "shift", "spark", "sphere", "spirit", "splendor",
	"surge", "sweep", "testament", "thrive", "timber", "timeless", "trace", "valor",
	"venture", "vista", "voyage", "wilderness", "zeal",
	"beam", "cadence", "chime", "echo", "fable", "feather", "gleam", "glimmer",
	"glint", "hymn", "luster", "mantle", "mural", "notion", "orchard", "patchwork",
	"pillar", "ripple", "solace", "sonnet", "taper", "tapestry", "tempo", "umbra",
	"verdant", "whisper", "wonder",
	"glyph", "herald", "lexicon", "logbook", "nexus", "oracle", "parable", "symbol",
	"token", "totem", "waypoint",
}

func randomPhrase() string {
	a := wordList[rand.Intn(len(wordList))]
	b := wordList[rand.Intn(len(wordList))]
	c := wordList[rand.Intn(len(wordList))]
	return a + "-" + b + "-" + c
}

func randomColor() string {
	return phraseColors[rand.Intn(len(phraseColors))]
}
