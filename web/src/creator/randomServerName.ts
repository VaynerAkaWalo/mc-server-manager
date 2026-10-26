const adjectives = [
  "Mossy", "Sleepy", "Ancient", "Glowing", "Frosty", "Hidden", "Rusty", "Lucky", "Dusty", "Sneaky",
  "Crimson", "Quiet", "Golden", "Copper", "Misty", "Leafy", "Cozy", "Snowy", "Sunny", "Stormy",
  "Brave", "Clever", "Curious", "Dizzy", "Fluffy", "Gentle", "Grumpy", "Happy", "Hungry", "Jolly",
  "Lazy", "Lonely", "Mighty", "Noisy", "Polished", "Proud", "Shiny", "Silent", "Silly", "Sparkly",
  "Speedy", "Spooky", "Sturdy", "Swift", "Tiny", "Wild", "Wobbly", "Wise", "Zesty", "Bouncy",
  "Cobbled", "Cracked", "Enchanted", "Floating", "Frozen", "Haunted", "Molten", "Mystic", "Sunken", "Twisted",
  "Withered", "Charred", "Blazing", "Echoing", "Weathered", "Oxidized", "Waxed", "Polar", "Chiseled", "Smooth",
  "Amber", "Azure", "Cyan", "Emerald", "Indigo", "Lime", "Magenta", "Scarlet", "Violet", "Ivory",
  "Arctic", "Desert", "Jungle", "Lunar", "Nether", "Ocean", "Savanna", "Swampy", "Taiga", "Tundra",
  "Feral", "Humble", "Noble", "Rowdy", "Sneezy", "Snoozy", "Tame", "Tipsy", "Toasty", "Witty",
]

const nouns = [
  "Creeper", "Axolotl", "Beacon", "Ravine", "Enderman", "Lantern", "Golem", "Biome", "Shulker", "Spawner",
  "Piglin", "Anvil", "Allay", "Amethyst", "Armadillo", "Bamboo", "Bat", "Bee", "Blaze", "Boat",
  "Bookshelf", "Bucket", "Cactus", "Camel", "Campfire", "Cat", "Cauldron", "Chest", "Chicken", "Cod",
  "Compass", "Cow", "Crafter", "Dolphin", "Donkey", "Dragon", "Elytra", "Ender", "Fox", "Frog",
  "Furnace", "Ghast", "Glowberry", "Goat", "Hoglin", "Hopper", "Horse", "Husk", "Iceberg", "Jukebox",
  "Kelp", "Ladder", "Lectern", "Llama", "Mangrove", "Minecart", "Mooshroom", "Mule", "Mushroom", "Ocelot",
  "Panda", "Parrot", "Phantom", "Pickaxe", "Pig", "Pillager", "Pufferfish", "Pumpkin", "Rabbit", "Redstone",
  "Salmon", "Sapling", "Scaffold", "Sculk", "Sheep", "Shovel", "Skeleton", "Slime", "Sniffer", "Snowman",
  "Spider", "Sponge", "Spruce", "Squid", "Strider", "Tadpole", "Torch", "Trident", "Turtle", "Villager",
  "Vindicator", "Warden", "Witch", "Wolf", "Zombie", "Cookie", "Melon", "Lodestone", "Obsidian", "Geode",
]

function pick(words: string[]): string {
  return words[Math.floor(Math.random() * words.length)]
}

export function randomServerName(): string {
  return `${pick(adjectives)} ${pick(nouns)}`
}
