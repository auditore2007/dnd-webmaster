package bestiary

// Формат: id|название|CR|тип|КД|HP|атака|скорость|СИЛ,ЛВК,ТЕЛ,ИНТ,МДР,ХАР|оружие|мультиатака|особенности|сопротивления|уязвимости|иммунитеты.
// Оружие: название:кубики:тип (знак ~ в конце — дальнобойное), через ";". Значения приведены по памяти SRD/Monster Manual: сверяйте с книгой.
const table = `
rat|Крыса|0|beast|10|1|0|20|2,11,9,2,10,4|Укус:1d2:piercing||Острый нюх|||
cat|Кошка|0|beast|12|2|0|40|3,15,10,3,12,7|Коготь:1d2:slashing||Острый нюх;Лазание|||
bat|Летучая мышь|0|beast|12|1|0|5|2,15,8,2,12,4|Укус:1d2:piercing||Эхолокация;Полёт|||
frog|Лягушка|0|beast|11|1|-1|20|1,13,8,1,8,3|Укус:1d2:piercing||Земноводное;Прыжок|||
raven|Ворон|0|beast|12|1|4|10|2,14,8,2,12,6|Клюв:1d2:piercing||Подражание;Полёт|||
spider|Паук|0|beast|12|1|4|20|2,14,8,1,10,2|Укус:1d2:piercing||Лазание по паутине;Паутинное чутьё|||
commoner|Обыватель|0|humanoid|10|4|2|30|10,10,10,10,10,10|Дубинка:1d4:bludgeoning|||||
giant-rat|Гигантская крыса|1/8|beast|12|7|4|30|7,15,11,2,10,4|Укус:1d4+2:piercing||Острый нюх;Тактика стаи|||
kobold|Кобольд|1/8|humanoid|12|5|4|30|7,15,9,8,7,8|Кинжал:1d4+2:piercing||Чувствительность к солнцу;Тактика стаи|||
bandit|Бандит|1/8|humanoid|12|11|3|30|11,12,12,10,10,10|Скимитар:1d6+1:slashing;Лёгкий арбалет:1d8+1:piercing~|||||
cultist|Культист|1/8|humanoid|12|9|3|30|11,12,10,10,11,10|Скимитар:1d6+1:slashing||Тёмная преданность|||
guard|Стражник|1/8|humanoid|16|11|3|30|13,12,12,10,11,10|Копьё:1d6+1:piercing|||||
noble|Дворянин|1/8|humanoid|15|9|2|30|11,12,11,12,14,16|Рапира:1d8:piercing|||||
tribal-warrior|Племенной воин|1/8|humanoid|12|11|3|30|13,11,12,8,11,8|Копьё:1d6+1:piercing||Тактика стаи|||
stirge|Кровосос|1/8|monstrosity|14|2|5|10|4,16,11,2,8,6|Кровососание:1d4+3:piercing||Полёт;Пьёт кровь|||
blood-hawk|Кровавый ястреб|1/8|beast|12|7|4|10|6,14,10,3,14,5|Клюв:1d4+2:piercing||Острое зрение;Полёт|||
giant-crab|Гигантский краб|1/8|beast|15|13|3|30|13,15,11,1,9,3|Клешня:1d6+1:bludgeoning||Земноводный|||
poison-snake|Ядовитая змея|1/8|beast|13|2|5|30|2,16,11,1,10,3|Укус:1d4+3:piercing||Слепое зрение;Яд|||
goblin|Гоблин|1/4|humanoid|15|7|4|30|8,14,10,10,8,8|Скимитар:1d6+2:slashing;Короткий лук:1d6+2:piercing~||Тёмное зрение;Проворный побег|||
skeleton|Скелет|1/4|undead|13|13|4|30|10,14,15,6,8,5|Короткий меч:1d6+2:piercing;Короткий лук:1d6+2:piercing~||Нежить||bludgeoning|poison
zombie|Зомби|1/4|undead|8|22|3|20|13,6,16,3,6,5|Удар:1d6+1:bludgeoning||Стойкость нежити|||poison
wolf|Волк|1/4|beast|13|11|4|40|12,15,12,3,12,6|Укус:2d4+2:piercing||Тактика стаи;Острый слух и нюх|||
acolyte|Аколит|1/4|humanoid|10|9|2|30|10,10,10,10,14,11|Дубинка:1d4:bludgeoning||Заклинатель-жрец|||
panther|Пантера|1/4|beast|12|13|4|50|14,15,10,3,14,7|Коготь:1d4+2:slashing;Укус:1d6+2:piercing||Наскок;Острый нюх|||
giant-frog|Гигантская лягушка|1/4|beast|11|18|3|30|12,13,11,2,10,3|Укус:1d6+1:piercing||Земноводная;Проглатывание|||
giant-wolf-spider|Гигантский паук-волк|1/4|beast|13|11|3|40|12,16,13,3,12,4|Укус:1d6+1:piercing||Паутинное чутьё;Лазание|||
pseudodragon|Псевдодракон|1/4|dragon|13|7|4|15|6,15,13,10,12,10|Укус:1d4+2:piercing||Острые чувства;Полёт;Магическая устойчивость|||
sprite|Спрайт|1/4|fey|15|2|6|10|3,18,10,14,13,11|Длинный меч:1d2:slashing;Короткий лук:1d2:piercing~||Полёт;Невидимость|||
pteranodon|Птеранодон|1/4|beast|13|13|5|10|12,15,10,2,9,5|Укус:2d4+2:piercing||Пикирование;Полёт|||
gnoll|Гнолл|1/2|humanoid|15|22|4|30|14,12,11,6,10,7|Копьё:1d6+2:piercing;Укус:1d4+2:piercing;Длинный лук:1d8+2:piercing~||Безумие плоти|||
hobgoblin|Хобгоблин|1/2|humanoid|18|11|3|30|13,12,12,10,10,9|Длинный меч:1d8+1:slashing;Длинный лук:1d8+1:piercing~||Военное превосходство|||
orc|Орк|1/2|humanoid|13|15|5|30|16,12,16,7,11,10|Большой топор:1d12+3:slashing;Метательное копьё:1d6+3:piercing||Агрессивный|||
scout|Разведчик|1/2|humanoid|13|16|4|30|11,14,12,11,13,11|Короткий меч:1d6+2:piercing;Длинный лук:1d8+2:piercing~|0,0|Острый слух;Острое зрение|||
thug|Головорез|1/2|humanoid|11|32|4|30|15,11,14,10,10,11|Булава:1d6+2:bludgeoning;Тяжёлый арбалет:1d10:piercing~|0,0|Тактика стаи|||
black-bear|Чёрный медведь|1/2|beast|11|19|3|40|15,10,14,2,12,7|Укус:1d6+2:piercing;Когти:2d4+2:slashing|0,1|Острый нюх;Лазание|||
lizardfolk|Ящеролюд|1/2|humanoid|15|22|4|30|15,10,13,7,12,7|Булава:1d6+2:bludgeoning;Метательное копьё:1d6+2:piercing||Задержка дыхания|||
shadow|Тень|1/2|undead|12|16|4|40|6,14,13,6,10,8|Холодное касание:2d6+2:necrotic||radiant,bludgeoning|acid,cold,fire,lightning,thunder|radiant|necrotic,poison
giant-wasp|Гигантская оса|1/2|beast|12|13|4|10|10,14,10,1,10,3|Жало:1d6+2:piercing||Полёт;Яд|||
crocodile|Крокодил|1/2|beast|12|19|4|20|15,10,13,2,10,5|Укус:1d10+2:piercing||Задержка дыхания|||
ape|Обезьяна|1/2|beast|12|19|5|30|16,14,14,6,12,7|Кулак:1d6+3:bludgeoning;Камень:1d6+3:bludgeoning~|0,0|Лазание|||
cockatrice|Василиск-петух|1/2|monstrosity|11|27|3|20|6,12,12,2,13,5|Укус:1d4+1:piercing||Окаменение|||
reef-shark|Рифовая акула|1/2|beast|12|22|4|40|14,13,13,1,10,4|Укус:1d8+2:piercing||Кровавое безумие|||
bugbear|Багбир|1|humanoid|16|27|4|30|15,14,13,8,11,9|Моргенштерн:2d8+2:piercing||Скрытная засада;Грубая сила|||
brown-bear|Бурый медведь|1|beast|11|34|5|40|19,10,16,2,13,7|Укус:1d8+4:piercing;Когти:2d6+4:slashing|0,1|Острый нюх|||
giant-spider|Гигантский паук|1|beast|14|26|5|30|14,16,12,2,11,4|Укус:1d8+3:piercing||Паутина;Лазание по паутине|||
dire-wolf|Лютый волк|1|beast|14|37|5|50|17,15,15,3,12,7|Укус:2d6+3:piercing||Тактика стаи;Острый слух и нюх|||
ghoul|Вурдалак|1|undead|12|22|4|30|13,15,10,7,10,6|Когти:2d4+2:slashing;Укус:2d6+2:piercing||Нежить;Паралич от когтей|||poison
lion|Лев|1|beast|12|26|5|50|17,15,13,3,12,8|Укус:1d8+3:piercing;Коготь:1d6+3:slashing|0,1|Прыжок;Острый нюх|||
tiger|Тигр|1|beast|12|37|5|40|17,15,14,3,12,8|Укус:1d10+3:piercing;Коготь:1d8+3:slashing|0,1|Прыжок;Острый нюх|||
imp|Имп|1|fiend|13|10|5|20|6,17,13,11,12,14|Жало:1d4+3:piercing||cold,fire,poison|||
harpy|Гарпия|1|monstrosity|11|38|3|20|12,13,12,7,10,13|Коготь:2d4+1:slashing;Дубина:1d4+1:bludgeoning|0,1|Чарующая песня;Полёт|||
spy|Шпион|1|humanoid|12|27|4|30|10,15,10,12,14,16|Короткий меч:1d6+2:piercing;Ручной арбалет:1d6+2:piercing~|0,0|Хитрое действие;Скрытая атака|||
specter|Призрак-спектр|1|undead|12|22|4|0|1,14,11,10,10,11|Высасывание жизни:3d6:necrotic||acid,cold,fire,lightning,thunder|||
animated-armor|Оживлённый доспех|1|construct|18|33|4|25|14,11,13,1,3,1|Удар:1d6+2:bludgeoning|0,0|Антимагия;Ложная внешность||poison|
dryad|Дриада|1|fey|11|22|2|30|10,12,11,14,15,18|Дубина:1d4:bludgeoning||Скрытность в дереве;Магическая устойчивость|||
giant-toad|Гигантская жаба|1|beast|11|39|4|20|15,13,13,2,10,3|Укус:1d10+2:piercing||Проглатывание;Земноводная|||
giant-boar|Гигантский кабан|2|beast|12|42|5|40|17,10,16,2,7,5|Бивни:2d6+3:slashing||Натиск;Безжалостный|||
ogre|Огр|2|giant|11|59|6|40|19,8,16,5,7,7|Палица:2d8+4:bludgeoning;Метательное копьё:2d6+4:piercing~|||||
gargoyle|Горгулья|2|elemental|15|52|4|30|15,11,16,6,11,7|Укус:1d6+2:piercing;Когти:1d6+2:slashing|0,1|Ложная внешность;Полёт|bludgeoning,piercing,slashing||poison
mimic|Мимик|2|monstrosity|12|58|5|15|17,12,15,5,13,8|Псевдопод:1d8+3:bludgeoning;Укус:1d8+3:piercing||Прилипание;Подражание||acid|
ghast|Гаст|2|undead|13|36|3|30|16,17,10,11,10,8|Когти:2d6+3:slashing;Укус:2d8+3:piercing||Зловоние;Паралич от когтей|necrotic||poison
polar-bear|Белый медведь|2|beast|12|42|7|40|20,10,16,2,13,7|Укус:1d8+5:piercing;Когти:2d6+5:slashing|0,1|Острый нюх|||
rhinoceros|Носорог|2|beast|11|45|7|40|21,8,15,2,12,6|Рог:2d8+5:bludgeoning||Натиск|||
allosaurus|Аллозавр|2|beast|13|51|6|60|19,13,17,2,12,5|Укус:2d10+4:piercing;Коготь:1d8+4:slashing||Наскок|||
centaur|Кентавр|2|monstrosity|12|45|6|50|18,14,14,9,13,11|Пика:1d6+4:piercing;Копыта:2d6+4:bludgeoning;Длинный лук:1d6+2:piercing~|0,1|Натиск|||
druid|Друид|2|humanoid|11|27|2|30|10,12,13,12,15,11|Боевой посох:1d6:bludgeoning||Друид-заклинатель|||
priest|Жрец|2|humanoid|13|27|2|30|10,10,12,13,16,13|Булава:1d6:bludgeoning||Божественный заклинатель|||
bandit-captain|Капитан бандитов|2|humanoid|15|65|5|30|15,16,14,14,11,14|Скимитар:1d6+3:slashing;Кинжал:1d4+3:piercing|0,0,1|Парирование|||
cult-fanatic|Фанатик культа|2|humanoid|13|33|4|30|11,14,12,10,13,14|Кинжал:1d4+2:piercing|0,0|Тёмная преданность;Заклинатель|||
ankheg|Анкхег|2|monstrosity|14|39|5|30|17,11,13,1,13,6|Укус:2d6+3:slashing||Рытьё;Кислотный плевок|||
gelatinous-cube|Студенистый куб|2|ooze|6|84|6|15|14,3,20,1,6,1|Ложноножка:3d6:acid||Поглощение;Прозрачный|||acid,lightning
constrictor-snake|Удав|2|beast|12|60|6|30|19,14,12,1,10,3|Укус:2d6+4:piercing;Сдавливание:2d8+4:bludgeoning|0,1|Слепое зрение|||
saber-toothed-tiger|Саблезубый тигр|2|beast|12|52|6|40|18,14,15,3,12,8|Укус:1d10+5:piercing;Коготь:2d6+5:slashing|0,1|Прыжок|||
will-o-wisp|Блуждающий огонёк|2|undead|19|22|4|50|1,28,10,13,14,11|Касание:2d8:lightning||Невидимость;Полёт;Бестелесный|acid,cold,fire,necrotic,thunder||lightning,poison
wererat|Крысолак|2|humanoid|12|60|4|30|10,15,12,11,10,8|Короткий меч:1d6+2:piercing;Ручной арбалет:1d6+2:piercing~|0,0|Ликантропия;Оборотень|||
owlbear|Совомедведь|3|monstrosity|13|59|7|40|20,12,17,3,12,7|Клюв:1d10+5:piercing;Когти:2d8+5:slashing|0,1|Острое зрение и нюх|||
minotaur|Минотавр|3|monstrosity|14|76|6|40|18,11,16,6,16,9|Большой топор:2d12+4:slashing;Бодание:2d8+4:piercing||Таранный натиск;Безрассудный|||
veteran|Ветеран|3|humanoid|17|58|5|30|16,13,14,10,11,10|Длинный меч:1d8+3:slashing;Короткий меч:1d6+3:piercing|0,0,1||||
knight|Рыцарь|3|humanoid|18|52|5|30|16,11,14,11,11,15|Двуручный меч:2d6+3:slashing;Тяжёлый арбалет:1d10:piercing~|0,0|Храбрый|||
werewolf|Волколак|3|humanoid|11|58|4|30|15,13,14,10,11,10|Укус:1d8+2:piercing;Когти:2d4+2:slashing;Копьё:1d8+2:piercing|0,1|Ликантропия;Острый нюх|||
mummy|Мумия|3|undead|11|58|5|20|16,8,15,6,10,12|Гнилой кулак:2d6+3:bludgeoning|0,0|Ужасающий взгляд;Гниль мумии||fire|poison
wight|Умертвие|3|undead|14|45|4|30|15,14,16,10,13,15|Длинный меч:1d8+2:slashing;Длинный лук:1d8+2:piercing~|0,0|Чувствительность к солнцу|necrotic||poison
manticore|Мантикора|3|monstrosity|14|68|5|30|17,16,17,7,12,8|Укус:1d8+3:piercing;Коготь:1d6+3:slashing;Шипы:1d8+3:piercing~|0,1,1|Полёт;Хвостовые шипы|||
doppelganger|Двойник|3|monstrosity|14|52|6|30|11,18,14,11,12,14|Удар:1d6+4:bludgeoning|0,0|Оборотень;Засада|||
basilisk|Василиск|3|monstrosity|15|52|5|20|16,8,15,2,8,7|Укус:2d6+3:piercing||Окаменяющий взгляд;Тёмное зрение|||
displacer-beast|Блуждающий зверь|3|monstrosity|13|85|6|40|18,15,16,6,12,8|Щупальце:1d6+4:bludgeoning;Коготь:2d4+4:slashing|0,0,1|Смещение|||
giant-scorpion|Гигантский скорпион|3|beast|15|52|4|40|15,13,15,1,9,3|Клешня:1d6+2:bludgeoning;Жало:1d10+2:piercing|0,0,1|Слепое зрение;Яд|||
green-hag|Зелёная карга|3|fey|17|82|6|30|18,12,16,13,14,14|Когти:2d8+4:slashing||Подражание;Заклинательница|||
hell-hound|Адская гончая|3|fiend|15|45|5|50|17,12,14,6,13,6|Укус:1d8+3:piercing||Огненное дыхание;Острый слух и нюх||fire|
phase-spider|Фазовый паук|3|monstrosity|13|32|6|30|15,15,12,6,10,6|Укус:1d10+2:piercing||Эфирный прыжок;Яд|||
nightmare|Кошмар|3|fiend|13|68|6|60|18,15,16,10,13,15|Копыта:2d8+4:bludgeoning||Огненный ореол;Эфирный бег||fire|
ankylosaurus|Анкилозавр|3|beast|15|68|7|30|19,11,15,2,12,5|Хвост:4d6+4:bludgeoning|||||
ettin|Эттин|4|giant|12|85|7|40|21,8,17,6,10,8|Боевой топор:2d8+5:slashing;Моргенштерн:2d8+5:piercing|0,1|Две головы|||
ghost|Привидение|4|undead|11|45|5|0|7,13,10,10,12,17|Иссушающее касание:4d6+3:necrotic||Бестелесный;Эфирный;Ужасный вид|acid,fire,lightning,thunder,bludgeoning,piercing,slashing||cold,necrotic,poison
banshee|Банши|4|undead|12|58|4|0|1,14,10,12,11,17|Коррупционное касание:3d6+2:necrotic||Бестелесная;Предсмертный вопль|acid,fire,lightning,thunder,bludgeoning,piercing,slashing||cold,necrotic,poison
black-pudding|Чёрный пудинг|4|ooze|7|85|5|20|16,5,16,1,6,1|Ложноножка:1d6+3:bludgeoning||Разделение;Едкий||acid,cold,lightning,slashing|
bulette|Булетта|5|monstrosity|17|94|7|40|19,11,21,2,10,5|Укус:4d12+4:piercing||Прыжок;Подземный бег|||
flameskull|Пылающий череп|4|undead|13|40|5|0|1,17,14,16,10,11|Огненный луч:3d6:fire||Полёт;Заклинатель|lightning,necrotic,piercing||cold,fire,poison
lamia|Ламия|4|monstrosity|13|97|5|30|16,13,15,14,15,16|Когти:2d10+3:slashing;Кинжал:2d4+3:piercing|0,1|Заклинательница|||
elephant|Слон|4|beast|12|76|8|40|22,9,17,3,11,6|Бивни:3d8+6:piercing;Топтание:3d10+6:bludgeoning||Топчущий натиск|||
succubus|Суккуб|4|fiend|15|66|5|30|8,17,13,15,12,20|Коготь:1d6+3:slashing||cold,fire,lightning,poison|||
wereboar|Вепрь-оборотень|4|humanoid|10|78|5|30|17,10,15,10,11,8|Бивни:2d6+3:slashing;Булава:2d6+3:bludgeoning||Ликантропия;Натиск|||
weretiger|Тигр-оборотень|4|humanoid|12|120|5|30|17,15,16,10,13,11|Когти:2d4+3:slashing;Укус:1d10+3:piercing|0,1|Ликантропия;Прыжок|||
troll|Тролль|5|giant|15|84|7|30|18,13,20,7,9,7|Укус:1d6+4:piercing;Коготь:2d6+4:slashing|0,1,1|Регенерация;Острый нюх|||
hill-giant|Холмовой великан|5|giant|13|105|8|40|21,8,19,5,9,6|Палица:3d8+5:bludgeoning;Бросок камня:3d10+5:bludgeoning~|0,0||||
air-elemental|Элементаль воздуха|5|elemental|15|90|8|0|14,20,14,6,10,6|Удар:2d8+5:bludgeoning|0,0|Вихрь;Полёт|lightning,thunder,bludgeoning,piercing,slashing||poison
fire-elemental|Элементаль огня|5|elemental|13|102|6|50|10,17,16,6,10,7|Касание:2d6+3:fire|0,0|Горение;Осветитель|bludgeoning,piercing,slashing||fire,poison
earth-elemental|Элементаль земли|5|elemental|17|126|8|30|20,8,20,5,10,5|Удар:2d8+5:bludgeoning|0,0|Скольжение по земле;Осадный монстр|bludgeoning,piercing,slashing|thunder|poison
water-elemental|Элементаль воды|5|elemental|14|114|7|30|18,14,18,5,10,8|Удар:2d8+4:bludgeoning|0,0|Замерзание;Вода|acid,bludgeoning,piercing,slashing||poison
flesh-golem|Плотяной голем|5|construct|9|93|7|30|19,9,18,6,10,5|Удар:2d8+4:bludgeoning|0,0|Ярость;Неизменный облик|bludgeoning,piercing,slashing||lightning,poison
gladiator|Гладиатор|5|humanoid|16|112|7|30|18,15,16,10,12,15|Копьё:2d6+4:piercing;Щит:2d4+4:bludgeoning|0,0,1|Храбрый;Яростный|||
vampire-spawn|Вампир-отпрыск|5|undead|15|82|6|30|16,16,11,11,10,12|Укус:2d6+3:necrotic;Когти:2d4+3:slashing|0,1|Регенерация;Паучье лазание|necrotic||
wraith|Вурдалак-призрак|5|undead|13|67|6|0|6,16,16,12,14,15|Касание жизни:4d8+3:necrotic||Бестелесный;Чувствительность к солнцу|acid,cold,fire,lightning,thunder|poison|
xorn|Ксорн|5|elemental|19|73|6|20|17,10,22,11,10,11|Коготь:1d6+3:slashing;Укус:3d6+3:piercing|0,0,0,1|Прохождение сквозь землю|piercing,slashing||
shambling-mound|Ползучий курган|5|plant|15|136|7|20|18,8,16,5,10,5|Удар:2d8+4:bludgeoning|0,0|Поглощение молнии|cold,fire||lightning
giant-crocodile|Гигантский крокодил|5|beast|14|85|8|30|21,9,17,2,10,7|Укус:3d10+5:piercing;Хвост:2d8+5:bludgeoning|0,1|Задержка дыхания|||
salamander|Огненная саламандра|5|elemental|15|90|7|30|18,14,15,11,10,12|Копьё:2d6+4:piercing;Хвост:2d6+4:bludgeoning|0,1|Горячее тело|bludgeoning,piercing,slashing|cold|fire
triceratops|Трицератопс|5|beast|13|95|9|50|22,9,17,2,11,5|Бодание:4d8+6:piercing;Топтание:3d10+6:bludgeoning||Топчущий натиск|||
wyvern|Виверна|6|dragon|13|110|7|20|19,10,16,5,12,6|Укус:2d6+4:piercing;Жало:2d6+4:piercing|0,1|Ядовитое жало;Полёт|||
mage|Маг|6|humanoid|12|40|5|30|9,14,11,17,12,11|Кинжал:1d4+2:piercing||Заклинатель-волшебник|||
medusa|Медуза|6|monstrosity|15|127|5|30|10,15,16,12,13,15|Змеиные волосы:1d6+2:piercing;Короткий меч:1d6+2:slashing;Длинный лук:1d8+2:piercing~|0,1|Окаменяющий взгляд|||
chimera|Химера|6|monstrosity|14|114|7|30|19,11,19,3,14,10|Укус:2d6+4:piercing;Рога:1d12+4:bludgeoning;Когти:2d6+4:slashing|0,1,2|Огненное дыхание;Полёт|||
vrock|Вррок|6|fiend|15|104|8|40|17,15,18,8,13,8|Клюв:2d6+3:piercing;Когти:2d10+3:slashing|0,1|Споры;Полёт|cold,fire,lightning,bludgeoning,piercing,slashing||poison
mammoth|Мамонт|6|beast|13|126|10|40|24,9,21,3,11,6|Бивни:4d8+7:piercing;Топтание:4d10+7:bludgeoning||Топчущий натиск|||
young-white-dragon|Молодой белый дракон|6|dragon|17|133|7|40|18,10,18,6,11,12|Укус:2d10+4:piercing;Коготь:2d6+4:slashing|0,1,1|Ледяное дыхание;Полёт|||cold
stone-giant|Каменный великан|7|giant|17|126|9|40|23,15,20,10,12,9|Дубина:3d8+6:bludgeoning;Бросок камня:4d10+6:bludgeoning~|0,0|Ловля камней|||
oni|Они|7|giant|16|110|7|30|19,11,16,14,12,15|Глефа:2d10+4:slashing|0,0|Магия;Оборотень;Регенерация|||
young-black-dragon|Молодой чёрный дракон|7|dragon|18|127|7|40|19,14,17,12,11,15|Укус:2d10+4:piercing;Коготь:2d6+4:slashing|0,1,1|Кислотное дыхание;Полёт|||acid
shield-guardian|Страж-щит|7|construct|17|142|7|30|18,8,18,7,10,3|Кулак:2d6+4:bludgeoning|0,0|Связь с амулетом;Поглощение заклинаний|||poison
young-green-dragon|Молодой зелёный дракон|8|dragon|18|136|7|40|19,12,17,16,13,15|Укус:2d10+4:piercing;Коготь:2d6+4:slashing|0,1,1|Ядовитое дыхание;Полёт|||poison
frost-giant|Ледяной великан|8|giant|15|138|9|40|23,9,21,9,10,12|Большой топор:3d12+6:slashing;Бросок камня:4d10+6:bludgeoning~|0,0||||cold
hydra|Гидра|8|monstrosity|15|172|8|30|20,12,20,2,10,7|Укус:1d10+5:piercing|0,0,0,0,0|Много голов;Задержка дыхания|||
tyrannosaurus|Тираннозавр|8|beast|13|136|10|50|25,10,19,2,12,9|Укус:4d12+7:piercing;Хвост:3d8+7:bludgeoning|0,1||||
assassin|Убийца|8|humanoid|15|78|6|30|11,16,14,13,11,10|Короткий меч:1d6+3:piercing;Лёгкий арбалет:1d8+3:piercing~|0,0|Скрытая атака;Убийство;Уклонение|poison||
cloaker|Плащевик|8|aberration|14|78|6|10|17,15,12,13,12,14|Хвост:3d6+3:slashing;Укус:2d6+3:piercing|0,1|Теневая маскировка;Полёт|||
fire-giant|Огненный великан|9|giant|18|162|11|30|25,9,23,10,14,13|Большой меч:6d6+7:slashing;Бросок камня:4d10+7:bludgeoning~|0,0||||fire
young-blue-dragon|Молодой синий дракон|9|dragon|18|152|9|40|21,10,19,14,13,17|Укус:2d10+5:piercing;Коготь:2d6+5:slashing|0,1,1|Молниеносное дыхание;Полёт|||lightning
clay-golem|Глиняный голем|9|construct|14|133|8|20|20,9,18,3,8,1|Удар:2d10+5:bludgeoning|0,0|Принятие урона;Берсерк|||acid,poison
cloud-giant|Облачный великан|9|giant|14|200|12|40|27,10,22,12,16,16|Моргенштерн:3d8+8:piercing;Бросок камня:4d10+8:bludgeoning~|0,0|Тонкий нюх|||
bone-devil|Костяной дьявол|9|fiend|19|142|8|40|18,16,18,13,14,16|Коготь:2d8+4:slashing;Жало:2d8+4:piercing|0,0,1|Дьявольское зрение|cold,bludgeoning,piercing,slashing|fire,poison|
young-red-dragon|Молодой красный дракон|10|dragon|18|178|10|40|23,10,21,14,11,19|Укус:2d10+6:piercing;Коготь:2d6+6:slashing|0,1,1|Тёмное зрение;Полёт;Огненное дыхание|||fire
stone-golem|Каменный голем|10|construct|17|178|10|30|22,9,20,3,11,1|Удар:3d8+6:bludgeoning|0,0|Замедляющий взгляд|||poison
aboleth|Аболет|10|aberration|17|135|9|10|21,9,15,18,15,18|Щупальце:2d6+5:bludgeoning|0,0,0|Слизь;Телепатия|||
deva|Дэва|10|celestial|17|136|8|30|18,18,18,17,20,20|Булава:1d6+4:bludgeoning|0,0|Ангельское оружие;Полёт;Магия|radiant,bludgeoning,piercing,slashing||
guardian-naga|Нага-страж|10|monstrosity|18|127|8|40|19,18,16,16,19,18|Укус:1d8+4:piercing||Плевок ядом;Заклинатель|||
roc|Рух|11|monstrosity|15|248|13|20|28,10,20,3,10,9|Клюв:4d8+9:piercing;Когти:4d6+9:slashing|0,1|Острое зрение;Полёт|||
djinni|Джинн|11|elemental|17|161|9|30|21,15,22,15,16,20|Скимитар:2d6+5:slashing|0,0,0|Вихрь;Полёт;Магия|||lightning,thunder
efreeti|Ифрит|11|elemental|17|200|10|40|22,12,24,16,15,16|Скимитар:2d6+6:slashing|0,0|Огненная форма;Полёт;Магия|||fire
behir|Бехир|11|monstrosity|17|168|10|50|23,16,18,7,14,12|Укус:3d10+6:piercing;Сдавливание:2d10+6:bludgeoning|0,1|Молниевое дыхание;Проглатывание|||lightning
remorhaz|Ремораз|11|monstrosity|17|195|11|30|24,13,21,4,10,5|Укус:6d10+7:piercing||Раскалённое тело;Проглатывание|||cold,fire
horned-devil|Рогатый дьявол|11|fiend|18|178|10|20|22,17,21,12,16,17|Вилы:2d8+6:piercing;Хвост:2d6+6:piercing|0,1|Полёт;Дьявольское зрение|cold,bludgeoning,piercing,slashing|fire,poison|
erinyes|Эриния|12|fiend|18|153|8|30|18,16,18,14,14,18|Длинный меч:2d8+4:slashing;Длинный лук:1d8+3:piercing~|0,0,0|Полёт;Дьявольское зрение|cold,bludgeoning,piercing,slashing|fire,poison|
young-gold-dragon|Молодой золотой дракон|10|dragon|18|178|10|40|23,14,21,16,13,20|Укус:2d10+6:piercing;Коготь:2d6+6:slashing|0,1,1|Огненное дыхание;Полёт|||fire
adult-white-dragon|Взрослый белый дракон|13|dragon|18|200|11|40|22,10,22,8,12,12|Укус:2d10+6:piercing;Коготь:2d6+6:slashing;Хвост:2d8+6:bludgeoning|0,1,1,2|Ледяное дыхание;Полёт;Страх|||cold
beholder|Бехолдер|13|aberration|18|180|5|0|10,14,18,17,15,17|Укус:4d6:piercing||Антимагический конус;Глазные лучи;Полёт|||
nalfeshnee|Налфешни|13|fiend|18|184|10|20|21,10,22,19,12,15|Укус:5d10+5:piercing;Коготь:3d6+5:slashing|0,1,1|Полёт;Страх|cold,fire,lightning,bludgeoning,piercing,slashing||poison
storm-giant|Штормовой великан|13|giant|16|230|14|50|29,14,20,16,18,18|Двуручный меч:6d6+9:slashing;Бросок камня:4d12+9:bludgeoning~|0,0|Дыхание под водой|cold||lightning,thunder
vampire|Вампир|13|undead|16|144|9|30|18,18,18,17,15,18|Когти:1d8+4:slashing;Укус:1d6+4:necrotic|0,1|Регенерация;Оборотень;Паучье лазание|necrotic,bludgeoning,piercing,slashing||
rakshasa|Ракшас|13|fiend|16|110|7|40|14,17,18,13,16,20|Коготь:2d6+2:slashing|0,0|Иммунитет к заклинаниям до 6 уровня|bludgeoning,piercing,slashing|radiant|
adult-black-dragon|Взрослый чёрный дракон|14|dragon|19|195|11|40|23,14,21,14,13,17|Укус:2d10+6:piercing;Коготь:2d6+6:slashing;Хвост:2d8+6:bludgeoning|0,1,1,2|Кислотное дыхание;Полёт;Страх|||acid
ice-devil|Ледяной дьявол|14|fiend|18|180|10|40|21,14,18,18,15,18|Укус:2d6+5:piercing;Коготь:2d4+5:slashing;Хвост:2d6+5:bludgeoning|0,1,2|Дьявольское зрение|bludgeoning,piercing,slashing|fire|cold,poison
adult-green-dragon|Взрослый зелёный дракон|15|dragon|19|207|11|40|23,12,21,18,15,17|Укус:2d10+6:piercing;Коготь:2d6+6:slashing;Хвост:2d8+6:bludgeoning|0,1,1,2|Ядовитое дыхание;Полёт;Страх|||poison
mummy-lord|Повелитель мумий|15|undead|17|97|9|20|18,10,17,11,18,16|Гнилой кулак:3d6+4:bludgeoning|0,0|Ужасающий взгляд;Заклинатель|necrotic,bludgeoning,piercing,slashing|fire|poison
purple-worm|Пурпурный червь|15|monstrosity|18|247|14|50|28,7,22,1,8,4|Укус:3d8+9:piercing;Жало:3d6+9:piercing|0,1|Прорыв;Проглатывание|||
iron-golem|Железный голем|16|construct|20|210|13|30|24,9,20,3,11,1|Меч:3d10+7:slashing;Удар:3d8+7:bludgeoning|0,0|Поглощение огня;Ядовитое дыхание|||fire,poison
marilith|Мариллит|16|fiend|18|189|9|40|18,20,20,18,16,20|Длинный меч:2d10+4:slashing;Хвост:2d10+4:bludgeoning|0,0,0,0,0,0,1|Магическое оружие;Реакция|cold,fire,lightning,bludgeoning,piercing,slashing||poison
planetar|Планетар|16|celestial|19|200|12|40|24,20,24,19,22,25|Двуручный меч:4d6+7:slashing|0,0|Ангельское оружие;Полёт;Магия|radiant,bludgeoning,piercing,slashing||
adult-red-dragon|Взрослый красный дракон|17|dragon|19|256|14|40|27,10,25,16,13,21|Укус:2d10+8:piercing;Коготь:2d6+8:slashing;Хвост:2d8+8:bludgeoning|0,1,1,2|Огненное дыхание;Полёт;Страх|||fire
death-knight|Рыцарь смерти|17|undead|20|180|11|30|20,11,20,12,16,18|Длинный меч:5d8+5:slashing|0,0,0|Магическая устойчивость;Адский огонь|||necrotic,poison
dragon-turtle|Дракон-черепаха|17|dragon|20|341|13|20|25,10,20,10,12,12|Укус:3d12+7:piercing;Коготь:2d8+7:slashing;Хвост:3d12+7:bludgeoning|0,1,1,2|fire|||
balor|Балор|19|fiend|19|262|14|40|26,15,22,20,16,22|Длинный меч:3d8+8:slashing;Кнут:2d6+8:slashing|0,1|Огненная смерть;Огненная аура;Полёт|cold,lightning,bludgeoning,piercing,slashing||fire,poison
pit-fiend|Пит-фиенд|20|fiend|19|300|14|30|26,14,24,22,18,24|Укус:4d6+8:piercing;Коготь:2d8+8:slashing;Булава:2d6+8:bludgeoning;Хвост:3d10+8:bludgeoning|0,1,2,3|Полёт;Дьявольское зрение;Аура страха|cold,bludgeoning,piercing,slashing||fire,poison
ancient-white-dragon|Древний белый дракон|20|dragon|20|333|14|40|26,10,26,10,13,14|Укус:2d10+8:piercing;Коготь:2d6+8:slashing;Хвост:2d8+8:bludgeoning|0,1,1,2|Ледяное дыхание;Полёт;Страх|||cold
lich|Лич|21|undead|17|135|12|30|11,16,16,20,14,16|Парализующее касание:3d6:cold||Заклинатель 18 уровня;Филактерия|cold,lightning,necrotic||poison
solar|Солар|21|celestial|21|243|15|50|26,22,26,25,25,30|Двуручный меч:4d6+8:slashing;Длинный лук:2d8+6:piercing~|0,0|Божественное оружие;Полёт;Исцеление|radiant,bludgeoning,piercing,slashing||necrotic,poison
ancient-black-dragon|Древний чёрный дракон|21|dragon|22|367|15|40|27,14,25,16,15,19|Укус:2d10+9:piercing;Коготь:2d6+9:slashing;Хвост:2d8+9:bludgeoning|0,1,1,2|Кислотное дыхание;Полёт;Страх|||acid
ancient-red-dragon|Древний красный дракон|24|dragon|22|546|17|40|30,10,29,18,15,23|Укус:4d10+10:piercing;Коготь:2d6+10:slashing;Хвост:2d8+10:bludgeoning|0,1,1,2|Огненное дыхание;Полёт;Страх|||fire
tarrasque|Тарраск|30|monstrosity|25|676|19|40|30,11,30,3,11,11|Укус:4d12+10:piercing;Коготь:4d8+10:slashing;Рог:4d10+10:piercing;Хвост:4d6+10:bludgeoning|0,1,1,2,3|Отражающий панцирь;Бессмертие|bludgeoning,piercing,slashing||fire,poison
`
