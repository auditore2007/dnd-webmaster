package rules

// Расы D&D 5e: основные из Книги игрока и популярные из Руководства Вуло, Мордикайнена и других книг.
// Описание и история — в lore.go. Особенности перечислены текстом; механика заложена у тех, что влияют на расчёты
// (бонусы характеристик, скорость, сопротивления, здоровье, «непоколебимая стойкость», дыхание дракона).
var dndRaces = []Race{
	{ID: "aarakocra", Name: "Ааракокра", Speed: 25, Bonus: bonus{"dex": 2, "wis": 1}, Traits: []string{"Полёт 50 футов", "Когти", "Воздушное чутьё"}},
	{ID: "aasimar", Name: "Аасимар", Speed: 30, Bonus: bonus{"cha": 2}, Resist: []string{"necrotic", "radiant"}, Traits: []string{"Тёмное зрение", "Небесное сопротивление", "Целительные руки", "Свет в руках"}, Subs: []Race{
		{ID: "protector", Name: "Защитник", Bonus: bonus{"wis": 1}, Traits: []string{"Сияющая душа: крылья и излучение"}},
		{ID: "scourge", Name: "Каратель", Bonus: bonus{"con": 1}, Traits: []string{"Сияющее поглощение: ожог вокруг"}},
		{ID: "fallen", Name: "Падший", Bonus: bonus{"str": 1}, Traits: []string{"Некротическая тень: ужас и крылья"}}}},
	{ID: "bugbear", Name: "Багбир", Speed: 30, Bonus: bonus{"str": 2, "dex": 1}, Traits: []string{"Тёмное зрение", "Длинные руки", "Могучее телосложение", "Скрытность", "Внезапная атака"}},
	{ID: "dragonborn", Name: "Драконорождённый", Speed: 30, Bonus: bonus{"str": 2, "cha": 1}, Traits: []string{"Драконье наследие", "Дыхание дракона", "Сопротивление урону"}},
	{ID: "dwarf", Name: "Дварф", Speed: 25, Bonus: bonus{"con": 2}, Resist: []string{"poison"}, Traits: []string{"Тёмное зрение", "Стойкость дварфов", "Знание камня"}, Subs: []Race{
		{ID: "hill", Name: "Холмовой", Bonus: bonus{"wis": 1}, HPLevel: 1, Traits: []string{"Дварфийская живучесть"}},
		{ID: "mountain", Name: "Горный", Bonus: bonus{"str": 2}, Traits: []string{"Обучение доспехам"}}}},
	{ID: "elf", Name: "Эльф", Speed: 30, Bonus: bonus{"dex": 2}, Traits: []string{"Тёмное зрение", "Обострённые чувства", "Наследие фей", "Транс"}, Subs: []Race{
		{ID: "high", Name: "Высший", Bonus: bonus{"int": 1}, Traits: []string{"Заговор волшебника"}},
		{ID: "wood", Name: "Лесной", Speed: 35, Bonus: bonus{"wis": 1}, Traits: []string{"Быстроногий", "Маскировка на природе"}},
		{ID: "drow", Name: "Тёмный (дроу)", Bonus: bonus{"cha": 1}, Traits: []string{"Превосходное тёмное зрение", "Чувствительность к солнцу", "Магия дроу"}}}},
	{ID: "firbolg", Name: "Фирболг", Speed: 30, Bonus: bonus{"wis": 2, "str": 1}, Traits: []string{"Магия фирболгов", "Скрытый шаг", "Могучее телосложение", "Речь зверей и листвы"}},
	{ID: "genasi", Name: "Дженази", Speed: 30, Bonus: bonus{"con": 2}, Subs: []Race{
		{ID: "air", Name: "Воздушный", Bonus: bonus{"dex": 1}, Traits: []string{"Бесконечное дыхание", "Левитация (раз за отдых)"}},
		{ID: "earth", Name: "Земляной", Bonus: bonus{"str": 1}, Traits: []string{"Хождение по земле", "Слияние с камнем"}},
		{ID: "fire", Name: "Огненный", Bonus: bonus{"int": 1}, Resist: []string{"fire"}, Traits: []string{"Тёмное зрение", "Сопротивление огню", "Пламя в руках"}},
		{ID: "water", Name: "Водный", Bonus: bonus{"wis": 1}, Resist: []string{"acid"}, Traits: []string{"Подводное дыхание", "Плавание 30 футов", "Сопротивление кислоте"}}}},
	{ID: "gnome", Name: "Гном", Speed: 25, Bonus: bonus{"int": 2}, Traits: []string{"Тёмное зрение", "Гномья хитрость"}, Subs: []Race{
		{ID: "forest", Name: "Лесной", Bonus: bonus{"dex": 1}, Traits: []string{"Природная иллюзия"}},
		{ID: "rock", Name: "Скальный", Bonus: bonus{"con": 1}, Traits: []string{"Знание изобретателя"}}}},
	{ID: "goblin", Name: "Гоблин", Speed: 30, Bonus: bonus{"dex": 2, "con": 1}, Traits: []string{"Тёмное зрение", "Ярость малых", "Проворный побег"}},
	{ID: "goliath", Name: "Голиаф", Speed: 30, Bonus: bonus{"str": 2, "con": 1}, Resist: []string{"cold"}, Traits: []string{"Прирождённый атлет", "Стойкость камня", "Могучее телосложение", "Горная выносливость"}},
	{ID: "halfelf", Name: "Полуэльф", Speed: 30, Bonus: bonus{"cha": 2, "dex": 1, "con": 1}, Traits: []string{"Тёмное зрение", "Наследие фей", "Универсальность навыков"}},
	{ID: "halforc", Name: "Полуорк", Speed: 30, Bonus: bonus{"str": 2, "con": 1}, Traits: []string{"Тёмное зрение", "Угрожающий вид", "Непоколебимая стойкость", "Свирепые атаки"}},
	{ID: "halfling", Name: "Полурослик", Speed: 25, Bonus: bonus{"dex": 2}, Traits: []string{"Удачливый", "Храбрый", "Проворство полурослика"}, Subs: []Race{
		{ID: "lightfoot", Name: "Легконогий", Bonus: bonus{"cha": 1}, Traits: []string{"Природная скрытность"}},
		{ID: "stout", Name: "Коренастый", Bonus: bonus{"con": 1}, Resist: []string{"poison"}, Traits: []string{"Стойкость коренастых"}}}},
	{ID: "hobgoblin", Name: "Хобгоблин", Speed: 30, Bonus: bonus{"con": 2, "int": 1}, Traits: []string{"Тёмное зрение", "Военная выучка", "Спасти лицо"}},
	{ID: "human", Name: "Человек", Speed: 30, Bonus: bonus{"str": 1, "dex": 1, "con": 1, "int": 1, "wis": 1, "cha": 1}, Traits: []string{"Универсальность"}},
	{ID: "kenku", Name: "Кенку", Speed: 30, Bonus: bonus{"dex": 2, "wis": 1}, Traits: []string{"Мастер подделок", "Подражание голосам", "Обучение навыкам"}},
	{ID: "kobold", Name: "Кобольд", Speed: 30, Bonus: bonus{"dex": 2, "str": -2}, Traits: []string{"Тёмное зрение", "Тактика стаи", "Чувствительность к солнцу", "Раболепие"}},
	{ID: "lizardfolk", Name: "Ящеролюд", Speed: 30, Bonus: bonus{"con": 2, "wis": 1}, Traits: []string{"Укус", "Природная броня", "Задержка дыхания", "Плавание 30 футов", "Охотничья мудрость"}},
	{ID: "minotaur", Name: "Минотавр", Speed: 30, Bonus: bonus{"str": 2, "con": 1}, Traits: []string{"Рога", "Таранный натиск", "Молот-рога", "Охотничье чутьё"}},
	{ID: "orc", Name: "Орк", Speed: 30, Bonus: bonus{"str": 2, "con": 1}, Traits: []string{"Тёмное зрение", "Агрессивность", "Могучее телосложение", "Первобытное чутьё"}},
	{ID: "tabaxi", Name: "Табакси", Speed: 30, Bonus: bonus{"dex": 2, "cha": 1}, Traits: []string{"Тёмное зрение", "Кошачья ловкость", "Кошачьи когти", "Лазание 20 футов", "Любопытство"}},
	{ID: "tiefling", Name: "Тифлинг", Speed: 30, Bonus: bonus{"cha": 2, "int": 1}, Resist: []string{"fire"}, Traits: []string{"Тёмное зрение", "Адское сопротивление", "Инфернальное наследие"}},
	{ID: "tortle", Name: "Тортл", Speed: 30, Bonus: bonus{"str": 2, "wis": 1}, Traits: []string{"Панцирь (КД 17)", "Когти", "Задержка дыхания", "Укрыться в панцире", "Чутьё выживания"}},
	{ID: "triton", Name: "Тритон", Speed: 30, Bonus: bonus{"str": 1, "con": 1, "cha": 1}, Resist: []string{"cold"}, Traits: []string{"Земноводный", "Плавание 30 футов", "Власть над водными тварями", "Страж глубин"}},
	{ID: "warforged", Name: "Воплощённый", Speed: 30, Bonus: bonus{"con": 2, "str": 1}, Resist: []string{"poison"}, Traits: []string{"Не нуждается во сне", "Композитный корпус (+1 КД)", "Стойкий", "Не ест и не дышит"}},
	{ID: "yuanti", Name: "Юань-ти чистокровный", Speed: 30, Bonus: bonus{"cha": 2, "int": 1}, Resist: []string{"poison"}, Traits: []string{"Тёмное зрение", "Врождённое колдовство", "Сопротивление магии", "Иммунитет к яду"}},
}
