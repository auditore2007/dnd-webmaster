package rules

// Таблица заклинаний D&D 5e (PHB, SRD, Xanathar, Tasha).
// Формат: id|название|уровень|школа|классы|описание[|режим|урон|тип|спасбросок|половина|область|за ячейку|заговор].
// Заклинания с режимом автоматизируются; остальные тратят ячейку, эффект решает мастер.
const spellTable = `bladeward|Защита от клинков|0|Ограждение|bard,sorcerer,wizard,warlock|Сопротивление физическому урону от оружия на 1 раунд.
chilltouch|Леденящее прикосновение|0|Некромантия|sorcerer,wizard,warlock|Призрачная рука: 1d8 некротического, цель не лечится до вашего следующего хода.|attack|1d8|necrotic|||||1
dancinglights|Танцующие огоньки|0|Воплощение|bard,sorcerer,wizard|До четырёх огоньков, парящих и перемещаемых бонусным действием.
druidcraft|Искусство друидов|0|Преобразование|druid|Мелкие природные чудеса: погода, цветок, искра.
friends|Дружба|0|Очарование|bard,sorcerer,warlock,wizard|Преимущество на Харизму против одного существа; потом оно понимает манипуляцию.
mending|Починка|0|Преобразование|bard,cleric,druid,sorcerer,wizard,artificer|Чинит один разрыв или поломку.
message|Сообщение|0|Преобразование|bard,sorcerer,wizard,artificer|Шёпотом передаёт сообщение существу в 120 футах.
minorillusion|Малая иллюзия|0|Иллюзия|bard,sorcerer,warlock,wizard|Звук или образ предмета до 5 футов.
produceflame|Сотворение пламени|0|Вызов|druid|Огонь в руке освещает; можно метнуть: 1d8 огнём.|attack|1d8|fire|||||1
resistance|Сопротивление|0|Ограждение|cleric,druid,artificer|Союзник добавляет 1d4 к одному спасброску.
shillelagh|Дубинка друида|0|Преобразование|druid|Дубина или посох бьют по Мудрости, 1d8.
sparethedying|Стабилизация|0|Некромантия|cleric,artificer|Умирающее существо стабилизируется.
thaumaturgy|Чудотворство|0|Преобразование|cleric|Малые знамения: дрожь земли, громовой голос, мигание огней.
truestrike|Верный удар|0|Прорицание|bard,sorcerer,warlock,wizard|Преимущество на следующую атаку по цели.
thornwhip|Терновый хлыст|0|Преобразование|druid,artificer|Лиана бьёт на 30 футов: 1d6 колющего, притягивает цель.|attack|1d6|piercing|||||1
tollthedead|Погребальный звон|0|Некромантия|cleric,wizard,warlock|Спасбросок Мудрости, иначе 1d8 некротического (1d12 если цель ранена).|save|1d8|necrotic|wis||||1
greenflameblade|Клинок зелёного пламени|0|Воплощение|sorcerer,wizard,warlock,artificer|Удар оружием с огненным ожогом соседней цели.
boomingblade|Грохочущий клинок|0|Воплощение|sorcerer,wizard,warlock,artificer|Удар оружием; если цель сдвинется, она получит грохот.
swordburst|Взрыв клинков|0|Колдовство|sorcerer,wizard,warlock,artificer|Мечи вокруг вас: 1d6 силой всем в 5 футах (спасбросок Ловкости).|save|1d6|force|dex||1||1
mindsliver|Расщепление разума|0|Очарование|sorcerer,warlock,wizard|1d6 психического и -1d4 к следующему спасброску цели.|save|1d6|psychic|int||||1
lightninglure|Манящая молния|0|Воплощение|sorcerer,wizard,warlock,artificer|Притягивает цель и бьёт 1d8 молнией.|save|1d8|lightning|str||||1
frostbite|Ледяной укус|0|Воплощение|druid,sorcerer,wizard,warlock,artificer|Спасбросок Телосложения, иначе 1d6 холода и помеха на атаку.|save|1d6|cold|con||||1
thunderclap|Громовой хлопок|0|Воплощение|bard,druid,sorcerer,wizard,warlock,artificer|Звуковой удар во всех в 5 футах: 1d6.|save|1d6|thunder|con||1||1
controlflames|Власть над огнём|0|Преобразование|druid,sorcerer,wizard|Двигает, гасит или меняет небольшое пламя.
createbonfire|Костёр|0|Вызов|druid,sorcerer,wizard,warlock,artificer|Огненный столб: 1d8 огня тем, кто в нём.
gust|Порыв|0|Преобразование|druid,sorcerer,wizard|Сдвигает существо или предмет порывом ветра.
moldearth|Лепка земли|0|Преобразование|druid,sorcerer,wizard|Двигает землю, создаёт мелкие формы.
shapewater|Власть над водой|0|Преобразование|druid,sorcerer,wizard|Двигает, замораживает или окрашивает воду.
primalsavagery|Дикий клык|0|Преобразование|druid|Зубы или когти: 1d10 кислоты ближним боем.|attack|1d10|acid|||||1
wordofradiance|Слово сияния|0|Воплощение|cleric|Вокруг вас вспышка: 1d6 излучением, спасбросок Телосложения.
alarm|Тревога|1|Ограждение|ranger,wizard,artificer|Ритуал: сигнал, если кто-то вторгнется в область.
animalfriendship|Дружба с животными|1|Очарование|bard,druid,ranger|Зверь становится дружелюбным.
armorofagathys|Доспех Агатиса|1|Ограждение|warlock|Временные HP и ледяной ответный урон.
armsofhadar|Руки Хадара|1|Вызов|warlock|Щупальца: 2d6 некротического и половина скорости.|save|2d6|necrotic|str|1|1|1d6|
bane|Порча|1|Очарование|bard,cleric|До трёх целей вычитают 1d4 из атак и спасбросков.
beastbond|Связь со зверем|1|Прорицание|druid,ranger|Телепатия со зверем и общее зрение.
causefear|Вселить страх|1|Некромантия|warlock,wizard|Существо в панике, спасбросок Мудрости.
colorspray|Цветной луч|1|Иллюзия|sorcerer,wizard|Ослепляет существ с самым малым запасом HP.
command|Приказ|1|Очарование|cleric,paladin,ranger|Одно слово приказа: стой, брось, беги.
compelledduel|Принуждение к дуэли|1|Очарование|paladin|Цель обязана сражаться с вами.
comprehendlanguages|Понимание языков|1|Прорицание|bard,sorcerer,warlock,wizard|Понимаете любой язык и письмо на час.
createordestroywater|Вода: создание или уничтожение|1|Преобразование|cleric,druid|10 галлонов воды или рассеять туман.
detectevilandgood|Обнаружение добра и зла|1|Прорицание|cleric,paladin|Чувствуете исчадий, нежить, небожителей.
detectpoison|Обнаружение яда и болезней|1|Прорицание|bard,cleric,druid,paladin,ranger|Видите яды, отравленные предметы.
disguiseself|Маскировка|1|Иллюзия|bard,sorcerer,wizard,artificer|Иной облик на час.
dissonantwhispers|Диссонансный шёпот|1|Очарование|bard|Спасбросок Мудрости: 3d6 психического и бегство.|save|3d6|psychic|wis|1||1d6|
divinefavor|Божественное благоволение|1|Воплощение|paladin|+1d4 излучения к удару оружием.
earthtremor|Дрожь земли|1|Воплощение|bard,druid,sorcerer|Дрожь: 1d6 дробящего всем в 10 футах и падение.|save|1d6|bludgeoning|dex||1|1d6|
ensnaringstrike|Опутывающий удар|1|Колдовство|ranger|Лозы опутывают цель после попадания.
entangle|Опутывание|1|Вызов|druid,ranger|Растения хватают существ в квадрате 20 футов.
expeditiousretreat|Поспешное отступление|1|Преобразование|sorcerer,warlock,wizard,artificer|Рывок бонусным действием.
faeriefire|Огонь фей|1|Воплощение|bard,druid,artificer|Цели светятся, по ним преимущество на атаку.
falselife|Ложная жизнь|1|Некромантия|sorcerer,wizard,artificer|Временные HP 1d4+4.
featherfall|Падение пёрышком|1|Преобразование|bard,sorcerer,wizard,artificer|Падающие замедляются.
findfamiliar|Поиск фамильяра|1|Вызов|wizard|Дух-спутник в облике зверька.
fogcloud|Туманное облако|1|Вызов|druid,ranger,sorcerer,wizard|Сфера тумана закрывает обзор.
goodberry|Чудо-ягоды|1|Преобразование|druid,ranger|Десять ягод лечат по 1 HP и кормят.
grease|Жир|1|Вызов|wizard,artificer|Скользкая поверхность: Ловкость или падение.
hailofthorns|Град шипов|1|Колдовство|ranger|После выстрела шипы: 1d10 колющего вокруг.
heroism|Героизм|1|Очарование|bard,paladin|Временные HP, иммунитет к страху.
hellishrebuke|Адское возмездие|1|Воплощение|warlock|Реакция: 2d10 огня обидчику.|save|2d10|fire|dex|1||1d10|
identify|Опознание|1|Прорицание|bard,wizard,artificer|Ритуал: узнаёте свойства предмета.
illusoryscript|Иллюзорный текст|1|Иллюзия|bard,warlock,wizard|Письмо видно только избранным.
jump|Прыжок|1|Преобразование|druid,ranger,sorcerer,wizard,artificer|Утраивает дальность прыжка.
longstrider|Скороход|1|Преобразование|bard,druid,ranger,wizard,artificer|+10 футов к скорости на час.
protectionevilgood|Защита от добра и зла|1|Ограждение|cleric,paladin,warlock,wizard|Преимущество на защиту от нежити, исчадий, фей.
purifyfood|Очищение пищи и питья|1|Преобразование|cleric,druid,paladin|Ритуал: очищает еду и воду.
rayofsickness|Луч болезни|1|Некромантия|sorcerer,wizard|Атака заклинанием: 2d8 яда и отравление.|attack|2d8|poison||||1d8|
sanctuary|Святилище|1|Ограждение|cleric,artificer|Не атакуют, пока вы не нападёте.
searingsmite|Палящая кара|1|Воплощение|paladin|+1d6 огня при ударе, поджог.
shieldoffaith|Щит веры|1|Ограждение|cleric,paladin|+2 КД существу на 10 минут.
silentimage|Безмолвный образ|1|Иллюзия|bard,sorcerer,wizard|Иллюзорный образ без звука.
speakwithanimals|Разговор со зверями|1|Прорицание|bard,druid,ranger|Понимаете и говорите со зверями.
thunderoussmite|Громовая кара|1|Воплощение|paladin|+2d6 звука и сбивание с ног.
unseenservant|Невидимый слуга|1|Вызов|bard,warlock,wizard|Невидимая сила выполняет мелкие дела.
witchbolt|Ведьмин снаряд|1|Воплощение|sorcerer,warlock,wizard|1d12 молнией и дуга.|attack|1d12|lightning||||1d12|
wrathfulsmite|Яростная кара|1|Воплощение|paladin|+1d6 психического и страх.
absorbelements|Поглощение стихий|1|Ограждение|druid,ranger,sorcerer,wizard,artificer|Реакция: сопротивление стихийному урону.
iceknife|Ледяной кинжал|1|Вызов|druid,sorcerer,wizard|Метает ледяной осколок: 1d10 колющего и взрыв.|attack|1d10|piercing|||||
zephyrstrike|Удар ветра|1|Преобразование|ranger|Скорость и преимущество на атаку.
aid|Помощь|2|Ограждение|cleric,paladin,artificer,ranger|Союзники получают +5 максимума HP.
alterself|Смена облика|2|Преобразование|sorcerer,wizard,artificer|Меняете облик, дышите водой или отращиваете оружие.
animalmessenger|Звериный посыльный|2|Очарование|bard,druid,ranger|Зверёк несёт весть.
arcanelock|Магический замок|2|Ограждение|wizard|Запирает дверь или сундук магией.
augury|Гадание|2|Прорицание|cleric,druid|Ритуал: предсказание исхода на ближайшие полчаса.
barkskin|Дубовая кожа|2|Преобразование|druid,ranger|КД не ниже 16.
blindnessdeafness|Слепота/глухота|2|Некромантия|bard,cleric,sorcerer,wizard|Цель слепнет или глохнет.
blur|Размытый образ|2|Иллюзия|sorcerer,wizard,artificer|Помеха на атаки по вам.
brandingsmite|Клеймящая кара|2|Воплощение|paladin|+2d6 излучения и видимость цели.
calmemotions|Усмирение чувств|2|Очарование|bard,cleric|Прекращает ярость и очарование.
continualflame|Вечное пламя|2|Воплощение|cleric,wizard,artificer|Светящийся огонь без жара.
cordonofarrows|Завеса стрел|2|Колдовство|ranger|Четыре стрелы бьют по входящим.
crownofmadness|Корона безумия|2|Очарование|bard,sorcerer,warlock,wizard|Существо атакует по вашему указанию.
darkness|Тьма|2|Воплощение|sorcerer,warlock,wizard|Магическая тьма, не видно даже с тёмным зрением.
darkvision|Тёмное зрение|2|Преобразование|druid,ranger,sorcerer,wizard,artificer|Видите в темноте на 60 футов.
detectthoughts|Чтение мыслей|2|Прорицание|bard,sorcerer,warlock,wizard|Читаете поверхностные мысли.
enhanceability|Улучшение характеристики|2|Преобразование|bard,cleric,druid,artificer|Преимущество на проверки одной характеристики.
enlargereduce|Увеличение/уменьшение|2|Преобразование|sorcerer,wizard,artificer|Меняет размер существа.
enthrall|Увлечение|2|Очарование|bard,warlock|Цели не замечают ничего, кроме вас.
findsteed|Поиск скакуна|2|Вызов|paladin|Небесный или феерический скакун.
findtraps|Поиск ловушек|2|Прорицание|cleric,druid,ranger|Чувствуете ловушки.
flameblade|Пылающий клинок|2|Воплощение|druid|Огненный меч: 3d6 огнём.|attack|3d6|fire|||||
flamingsphere|Пылающая сфера|2|Вызов|druid,wizard|Катится огненный шар: 2d6 огнём в квадрате.|save|2d6|fire|dex|1|1|1d6|
gentlerepose|Мирный покой|2|Некромантия|cleric,wizard|Тело не разлагается.
gustofwind|Порыв ветра|2|Воплощение|druid,sorcerer,wizard|Линия сильного ветра сдвигает существ.
heatmetal|Раскалённый металл|2|Преобразование|bard,druid,artificer|Металл обжигает: 2d8 огня.|auto|2d8|fire||||1d8|
knock|Открывание|2|Преобразование|bard,sorcerer,wizard|Открывает запертое.
lesserrestoration|Малое восстановление|2|Некромантия|bard,cleric,druid,paladin,artificer|Снимает болезнь или состояние.
levitate|Левитация|2|Преобразование|sorcerer,wizard,artificer|Цель парит на 20 футов.
locateobject|Поиск предмета|2|Прорицание|bard,cleric,druid,paladin,ranger,wizard,artificer|Чувствуете направление к предмету.
magicmouth|Волшебные уста|2|Иллюзия|bard,wizard|Рот на поверхности произносит сообщение.
magicweapon|Магическое оружие|2|Преобразование|paladin,wizard,artificer|Оружие становится +1.
melfsacidarrow|Кислотная стрела Мельфа|2|Вызов|wizard|Атака заклинанием: 4d4 кислоты.|attack|4d4|acid||||1d4|
mirrorimage|Зеркальное отражение|2|Иллюзия|sorcerer,warlock,wizard|Три двойника сбивают атаки.
moonbeam|Лунный луч|2|Воплощение|druid|Столб света: 2d10 излучения в квадрате.|save|2d10|radiant|con|1|1|1d10|
passwithouttrace|Бесследное передвижение|2|Ограждение|druid,ranger|+10 к скрытности для группы.
prayerofhealing|Молитва исцеления|2|Воплощение|cleric,paladin|Группа лечится на 2d8 + мод.|heal|2d8|||||1d8|
rayofenfeeblement|Луч слабости|2|Некромантия|warlock,wizard|Цель слабеет, урон от неё снижен.
ropetrick|Верёвочный фокус|2|Преобразование|ranger,wizard,artificer|Верёвка ведёт в убежище.
seeinvisibility|Видение невидимого|2|Прорицание|bard,sorcerer,wizard,artificer|Видите невидимое.
shatter|Раскол|2|Воплощение|bard,sorcerer,warlock,wizard|Звук: 3d8 в сфере 10 футов, Телосложение.|save|3d8|thunder|con|1|1|1d8|
silence|Тишина|2|Иллюзия|bard,cleric,ranger|Сфера без звука блокирует заклинания.
spiderclimb|Паучье лазание|2|Преобразование|sorcerer,warlock,wizard,artificer|Лазание по стенам.
spikegrowth|Рост шипов|2|Преобразование|druid,ranger|Земля в шипах: 2d4 за каждые 5 футов.
suggestion|Внушение|2|Очарование|bard,sorcerer,warlock,wizard|Внушённое действие звучит разумно.
web|Паутина|2|Вызов|sorcerer,wizard,artificer|Липкая паутина, Ловкость или опутаны.
zoneoftruth|Зона правды|2|Очарование|bard,cleric,paladin|Цели не могут сказать явную ложь.
dragonsbreath|Дыхание дракона|2|Преобразование|sorcerer,wizard|Цель дышит стихией: 3d6 урона.|save|3d6|fire|dex|1|1|1d6|
mindspike|Разумный шип|2|Прорицание|sorcerer,warlock,wizard|3d8 психического и видите цель.|save|3d8|psychic|wis|1||1d8|
pyrotechnics|Пиротехника|2|Преобразование|bard,sorcerer,wizard,artificer|Огонь в вспышку или дым.
aganazzarsscorcher|Жгучий луч Аганаццара|2|Воплощение|sorcerer,wizard|Линия огня: 3d8, Ловкость.|save|3d8|fire|dex|1|1|1d8|
healingspirit|Целительный дух|2|Вызов|druid,ranger|Дух лечит 1d6 + мод каждый ход.
animatedead|Оживление мёртвых|3|Некромантия|cleric,warlock,wizard|Поднимает мертвеца скелетом или зомби.
beaconofhope|Маяк надежды|3|Ограждение|cleric|Преимущество на спасброски, максимум лечения.
bestowcurse|Проклятие|3|Некромантия|bard,cleric,wizard|Накладывает проклятие на цель.
blink|Мерцание|3|Преобразование|sorcerer,wizard,artificer|Исчезаете в эфир каждый ход.
calllightning|Призыв молнии|3|Вызов|druid|Туча бьёт молниями.|save|3d10|lightning|dex|1|1|1d10|
clairvoyance|Ясновидение|3|Прорицание|bard,cleric,sorcerer,wizard|Сенсор в известном месте.
conjureanimals|Призыв зверей|3|Вызов|druid,ranger|Рой зверей служит вам.
conjurebarrage|Шквал снарядов|3|Вызов|ranger|Дождь стрел или метательных снарядов.
createfood|Создание пищи и воды|3|Вызов|cleric,paladin,artificer|Еда для 15 существ на день.
daylight|Дневной свет|3|Воплощение|cleric,druid,paladin,ranger,sorcerer|Сфера яркого света на час.
elementalweapon|Стихийное оружие|3|Преобразование|druid,paladin,artificer|+1d4 урона стихией.
fear|Страх|3|Иллюзия|bard,sorcerer,warlock,wizard|Конус страха, бегство.
feigndeath|Притворная смерть|3|Некромантия|bard,cleric,druid,wizard|Ритуал: подобие смерти.
gaseousform|Газообразная форма|3|Преобразование|sorcerer,warlock,wizard|Превращаете цель в облако.
glyphofwarding|Охранный символ|3|Ограждение|bard,cleric,wizard,artificer|Охранная руна взрывается или налагает заклинание.
hungerofhadar|Голод Хадара|3|Вызов|warlock|Сфера пустоты: 2d6 холода и 2d6 кислоты.
hypnoticpattern|Гипнотический узор|3|Иллюзия|bard,sorcerer,warlock,wizard|Цели очарованы и обездвижены.
lightningarrow|Стрела молнии|3|Преобразование|ranger|Следующая стрела бьёт 4d8 молнией.
tinyhut|Хижина Леомунда|3|Воплощение|bard,wizard|Ритуал: купол укрытия на 8 часов.
lifetransference|Передача жизни|3|Некромантия|cleric,wizard|Ваша жизнь лечит другого.
magiccircle|Магический круг|3|Ограждение|cleric,paladin,warlock,wizard|Круг против заданного типа существ.
majorimage|Большой образ|3|Иллюзия|bard,sorcerer,warlock,wizard|Иллюзия со звуком, запахом и движением.
masshealingword|Массовое лечебное слово|3|Воплощение|cleric|До шести целей лечатся на 1d4 + мод.|heal|1d4||||1|1d4|
meldintostone|Слияние с камнем|3|Преобразование|cleric,druid,ranger|Скрываетесь в камне.
nondetection|Необнаружимость|3|Ограждение|bard,druid,ranger,wizard|Нельзя найти магией.
phantomsteed|Призрачный скакун|3|Иллюзия|wizard|Лошадь-призрак на час.
plantgrowth|Рост растений|3|Преобразование|bard,druid,ranger|Растения разрастаются или дают урожай.
protectionenergy|Защита от энергии|3|Ограждение|cleric,druid,ranger,sorcerer,wizard,artificer|Сопротивление одному стихийному типу.
removecurse|Снятие проклятия|3|Ограждение|cleric,paladin,warlock,wizard|Снимает проклятия.
sending|Послание|3|Воплощение|bard,cleric,wizard|25 слов получателю на любом расстоянии.
sleetstorm|Град и ветер|3|Вызов|druid,sorcerer,wizard|Сильный град делает землю скользкой.
slow|Замедление|3|Преобразование|sorcerer,wizard|До шести существ замедлено.
speakdead|Разговор с мёртвыми|3|Некромантия|bard,cleric|Труп отвечает на пять вопросов.
spiritguardians|Духи-хранители|3|Вызов|cleric|Духи кружат: 3d8 излучения или некротического.|save|3d8|radiant|wis|1|1|1d8|
stinkingcloud|Зловонное облако|3|Вызов|bard,sorcerer,wizard|Облако тошноты.
tongues|Языки|3|Прорицание|bard,cleric,sorcerer,warlock,wizard|Понимаете и говорите на любом языке.
vampirictouch|Вампирическое касание|3|Некромантия|warlock,wizard|Касание: 3d6 некротического и вы лечитесь.|attack|3d6|necrotic||||1d6|
waterbreathing|Подводное дыхание|3|Преобразование|druid,ranger,sorcerer,wizard,artificer|Ритуал: до десяти существ дышат под водой.
waterwalk|Хождение по воде|3|Преобразование|cleric,druid,ranger,sorcerer,artificer|Ритуал: ходьба по жидкости.
windwall|Стена ветра|3|Воплощение|druid,ranger|Ветер отбивает стрелы и газы.
auraofvitality|Аура жизненности|3|Воплощение|cleric,druid,paladin|Бонусным действием лечит 2d6.|heal|2d6||||||
blindingsmite|Ослепляющая кара|3|Воплощение|paladin|+3d8 излучения и слепота.
crusadersmantle|Плащ крестоносца|3|Воплощение|paladin|+1d4 излучения вам и союзникам.
eruptingearth|Извержение земли|3|Преобразование|druid,sorcerer,wizard|Земля вспучивается: 3d12 дробящего.|save|3d12|bludgeoning|dex|1|1|1d12|
flamearrows|Огненные стрелы|3|Преобразование|druid,ranger,wizard,artificer|Стрелы жгут ещё: 1d6 огнём.
melfsminutemeteors|Метеоры Мельфа|3|Воплощение|sorcerer,wizard|Шесть метеоров по 2d6 огня.
thunderstep|Громовой шаг|3|Колдовство|sorcerer,warlock,wizard|Телепорт с ударом грома 3d10.|save|3d10|thunder|con|1|1|1d10|
wallofsand|Стена песка|3|Воплощение|wizard|Стена из песка блокирует обзор.
wallofwater|Стена воды|3|Воплощение|druid,sorcerer,wizard|Стена воды ослабляет огонь.
arcaneeye|Волшебный глаз|4|Прорицание|wizard,artificer|Невидимый глаз-разведчик.
banishment|Изгнание|4|Ограждение|cleric,paladin,warlock,sorcerer,wizard|Выбрасывает цель в иной план.
evardsblacktentacles|Чёрные щупальца Эварда|4|Вызов|wizard|Щупальца хватают: 3d6 дробящего.|save|3d6|bludgeoning|dex||1||
blight|Увядание|4|Некромантия|druid,sorcerer,warlock,wizard|8d8 некротического, Телосложение.|save|8d8|necrotic|con|1||1d8|
compulsion|Принуждение|4|Очарование|bard|Существа бредут туда, куда прикажете.
confusion|Смятение|4|Очарование|bard,druid,sorcerer,wizard|Существа делают случайные действия.
conjureminorelementals|Призыв малых элементалей|4|Вызов|druid,wizard|Элементали служат вам.
conjurewoodlandbeings|Призыв лесных существ|4|Вызов|druid,ranger|Феи служат вам.
controlwater|Власть над водой|4|Преобразование|cleric,druid,wizard|Двигает водные массы.
deathward|Защита от смерти|4|Ограждение|cleric,paladin|Спасает от падения до 0 HP один раз.
dimensiondoor|Переход между измерениями|4|Колдовство|bard,sorcerer,warlock,wizard|Телепорт на 500 футов.
divination|Гадание|4|Прорицание|cleric,druid,wizard|Ритуал: знамение от бога.
dominatebeast|Подчинение зверя|4|Очарование|druid,ranger,sorcerer|Зверь подчиняется.
fabricate|Изготовление|4|Преобразование|wizard,artificer|Мгновенно изготавливает предметы.
fireshield|Огненный щит|4|Воплощение|wizard|Пламя или холод ранит нападающих.
freedomofmovement|Свобода движения|4|Ограждение|bard,cleric,druid,ranger,artificer|Нельзя замедлить или опутать.
giantinsect|Гигантские насекомые|4|Преобразование|druid|Превращает насекомых в огромных.
greaterinvisibility|Высшая невидимость|4|Иллюзия|bard,sorcerer,wizard|Невидимость даже при атаке.
guardianoffaith|Страж веры|4|Вызов|cleric|Небесный страж: 20 урона нападающим.
hallucinatoryterrain|Иллюзорный ландшафт|4|Иллюзия|bard,druid,warlock,wizard|Меняет вид местности.
icestorm|Ледяная буря|4|Воплощение|druid,sorcerer,wizard|Град: 2d8 дробящего и 4d6 холода.|save|4d6|cold|dex|1|1|1d6|
secretchest|Тайный сундук Леомунда|4|Преобразование|wizard|Сундук в эфире.
locatecreature|Поиск существа|4|Прорицание|bard,cleric,druid,paladin,ranger,wizard|Чувствуете путь к существу.
faithfulhound|Верный пёс Мордекайнена|4|Вызов|wizard|Невидимый пёс охраняет.
privatesanctum|Тайное убежище Мордекайнена|4|Ограждение|wizard|Защита помещения от всего.
resilientsphere|Стойкая сфера Отилюка|4|Воплощение|wizard|Заключает существо в сферу.
phantasmalkiller|Фантомный убийца|4|Иллюзия|wizard|Кошмар: 4d10 психического каждый ход.|save|4d10|psychic|wis|||1d10|
polymorph|Превращение|4|Преобразование|bard,druid,sorcerer,wizard|Превращает существо в зверя.
stoneshape|Придание камню формы|4|Преобразование|cleric,druid,wizard,artificer|Лепка камня.
stoneskin|Каменная кожа|4|Ограждение|druid,ranger,sorcerer,wizard,artificer|Сопротивление немагическому оружию.
walloffire|Огненная стена|4|Воплощение|druid,sorcerer,wizard|Стена: 5d8 огня.|save|5d8|fire|dex|1|1|1d8|
auraoflife|Аура жизни|4|Ограждение|cleric,paladin|Сопротивление некротическому, нельзя уменьшить максимум HP.
banishingsmite|Изгоняющая кара|4|Ограждение|paladin|+5d10 силой, изгнание.
staggeringsmite|Ошеломляющая кара|4|Воплощение|paladin|+4d6 психического и помеха.
charmmonster|Очарование монстра|4|Очарование|bard,druid,sorcerer,warlock,wizard|Очаровывает любое существо.
sickeningradiance|Тошнотворное сияние|4|Воплощение|sorcerer,warlock,wizard|Свет: 4d10 излучения и истощение.
stormsphere|Грозовая сфера|4|Воплощение|sorcerer,wizard|Шторм-шар: 2d6 звука и 4d6 молнии.|save|2d6|bludgeoning|str||1|1d6|
wateryspheres|Водяная сфера|4|Вызов|druid,sorcerer,wizard|Захватывает существ в пузыре.
animateobjects|Оживление предметов|5|Преобразование|bard,sorcerer,wizard,artificer|Предметы дерутся за вас.
antilifeshell|Оболочка против жизни|5|Ограждение|druid|Преграда для живых существ.
awaken|Пробуждение|5|Преобразование|bard,druid|Зверь или дерево обретает разум.
bigbyshand|Рука Бигби|5|Воплощение|wizard,artificer|Огромная рука бьёт, хватает или защищает.
cloudkill|Облако смерти|5|Вызов|sorcerer,wizard|Ядовитое облако: 5d8 яда.|save|5d8|poison|con|1|1|1d8|
commune|Общение|5|Прорицание|cleric|Ритуал: три вопроса божеству.
communewithnature|Общение с природой|5|Прорицание|druid,ranger|Ритуал: знание о местности.
coneofcold|Конус холода|5|Воплощение|sorcerer,wizard|60 футов, 8d8 холода, Телосложение.|save|8d8|cold|con|1|1|1d8|
conjureelemental|Призыв элементаля|5|Вызов|druid,wizard|Элементаль служит вам.
conjurevolley|Град стрел|5|Вызов|ranger|Дождь стрел: 8d8.
contactplane|Контакт с иным планом|5|Прорицание|warlock,wizard|Ритуал: вопросы потусторонним.
contagion|Заражение|5|Некромантия|cleric,druid|Болезнь на цели.
creation|Создание|5|Иллюзия|sorcerer,wizard|Создаёт предмет из теневой материи.
disperseevil|Рассеивание зла и добра|5|Ограждение|cleric,paladin|Защита от небесных, исчадий и нежити.
dominateperson|Подчинение личности|5|Очарование|bard,sorcerer,wizard|Подчиняет гуманоида.
dream|Сон|5|Иллюзия|bard,warlock,wizard|Посещаете чужие сны.
flamestrike|Огненный столп|5|Воплощение|cleric|Столп: 4d6 огня и 4d6 излучения.|save|8d6|fire|dex|1|1|1d6|
geas|Обет|5|Очарование|bard,cleric,druid,paladin,wizard|Принуждение выполнить задачу.
greaterrestoration|Высшее восстановление|5|Некромантия|bard,cleric,druid,artificer|Снимает тяжёлые состояния.
hallow|Освящение|5|Воплощение|cleric|Освящает область.
holdmonster|Удержание монстра|5|Очарование|bard,warlock,sorcerer,wizard|Парализует любое существо.
insectplague|Нашествие насекомых|5|Вызов|cleric,druid,sorcerer|Рой: 4d10 колющего.|save|4d10|piercing|con|1|1|1d10|
legendlore|Знание легенд|5|Прорицание|bard,cleric,wizard|Узнаёте историю известного.
masscurewounds|Массовое лечение ран|5|Воплощение|bard,cleric,druid|До шести целей лечатся на 3d8 + мод.|heal|3d8||||1|1d8|
mislead|Обман|5|Иллюзия|bard,warlock|Создаёте двойника.
modifymemory|Изменение памяти|5|Очарование|bard,wizard|Меняет воспоминания цели.
passwall|Проход в стене|5|Преобразование|wizard|Проём в стене.
planarbinding|Привязка к плану|5|Ограждение|bard,cleric,druid,warlock,wizard|Привязывает призванное существо.
raisedead|Воскрешение мёртвых|5|Некромантия|bard,cleric,paladin|Возвращает к жизни до 10 дней.
telepathicbond|Телепатическая связь Рэри|5|Прорицание|bard,wizard|Мысленная связь союзников.
reincarnate|Реинкарнация|5|Некромантия|druid|Новое тело для души.
scrying|Наблюдение|5|Прорицание|bard,cleric,druid,warlock,wizard|Наблюдаете за целью.
seeming|Личина|5|Иллюзия|bard,sorcerer,wizard|Меняет вид группы.
telekinesis|Телекинез|5|Преобразование|sorcerer,wizard|Двигаете существ и предметы разумом.
teleportationcircle|Круг телепортации|5|Колдовство|bard,sorcerer,wizard|Постоянный круг.
treestride|Шаг по деревьям|5|Колдовство|druid,ranger|Из дерева в дерево.
wallofforce|Силовая стена|5|Воплощение|wizard|Непроницаемая стена.
wallofstone|Каменная стена|5|Воплощение|cleric,druid,sorcerer,wizard,artificer|Стена из камня.
destructivewave|Разрушительная волна|5|Воплощение|paladin|Волна: 5d6 звука и 5d6 излучения.|save|5d6|thunder|con|1|1||
circleofpower|Круг силы|5|Ограждение|paladin|Союзники сильны к магии.
swiftquiver|Стремительный колчан|5|Преобразование|ranger|Колчан сам вылетает.
steelwindstrike|Удар стального ветра|5|Колдовство|ranger,wizard|Телепорт и 6d10 урона пяти целям.|attack|6d10|force|||||
synapticstatic|Синаптический шок|5|Очарование|bard,sorcerer,warlock,wizard|8d6 психического и смятение.|save|8d6|psychic|int|1|1||
enervation|Вытягивание сил|5|Некромантия|sorcerer,warlock,wizard|Высасывает жизнь: 4d8.|save|4d8|necrotic|dex|||1d8|
dansemacabre|Пляска мёртвых|5|Некромантия|warlock,wizard|Поднимает до пяти нежитей.
immolation|Испепеление|5|Воплощение|sorcerer,wizard|Горящая цель: 8d6 огня.|save|8d6|fire|dex|1|||
holyweapon|Священное оружие|5|Воплощение|cleric,paladin|Оружие светит и взрывается.
controlwinds|Власть над ветром|5|Преобразование|druid,sorcerer,wizard|Управляете ветром.
maelstrom|Водоворот|5|Воплощение|druid|Водоворот: 6d6 дробящего.|save|6d8|bludgeoning|str||1||
transmuterock|Преобразование камня|5|Преобразование|druid,wizard,artificer|Камень в грязь или обратно.
bladebarrier|Барьер из клинков|6|Воплощение|cleric|Стена: 6d10 рубящего.|save|6d10|slashing|dex|1|1||
chainlightning|Цепная молния|6|Воплощение|sorcerer,wizard|10d8 молнией основной цели и дуги.|save|10d8|lightning|dex|1|1||
circleofdeath|Круг смерти|6|Некромантия|sorcerer,warlock,wizard|Сфера: 8d6 некротического.|save|8d6|necrotic|con|1|1|2d6|
conjurefey|Призыв феи|6|Вызов|druid,warlock|Фея служит вам.
contingency|Условное заклинание|6|Воплощение|wizard|Заклинание срабатывает по условию.
createundead|Создание нежити|6|Некромантия|cleric,warlock,wizard|Гули, вурдалаки, призраки.
disintegrate|Дезинтеграция|6|Преобразование|sorcerer,wizard|10d6+40 силой, Ловкость.|save|10d6+40|force|dex|||3d6|
instantsummons|Мгновенный вызов Дроумиджа|6|Вызов|wizard|Вызывает ваш предмет в руку.
eyebite|Глазной укус|6|Некромантия|bard,sorcerer,warlock,wizard|Взгляд усыпляет, пугает или заражает.
findthepath|Поиск пути|6|Прорицание|bard,cleric,druid|Знаете кратчайший путь.
fleshtostone|Обращение в камень|6|Преобразование|warlock,wizard|Каменеет, Телосложение.
forbiddance|Запрет|6|Ограждение|cleric|Запрет на телепортацию и вход.
globeofinvulnerability|Сфера неуязвимости|6|Ограждение|sorcerer,wizard|Блокирует заклинания до 5 уровня.
guardsandwards|Стражи и обереги|6|Ограждение|bard,wizard|Защита здания.
harm|Вред|6|Некромантия|cleric|14d6 некротического, Телосложение.|save|14d6|necrotic|con|1|||
heal|Исцеление|6|Воплощение|cleric,druid|Лечит 70 HP и снимает хвори.|heal|20d6||||||
heroesfeast|Пир героев|6|Вызов|bard,cleric|Пир даёт иммунитет и максимум HP.
magicjar|Магический сосуд|6|Некромантия|wizard|Перенос души в сосуд.
masssuggestion|Массовое внушение|6|Очарование|bard,sorcerer,warlock,wizard|Внушение для 12 существ.
moveearth|Перемещение земли|6|Преобразование|druid,sorcerer,wizard|Двигает землю.
freezingsphere|Ледяная сфера Отилюка|6|Воплощение|wizard|10d6 холода по области.|save|10d6|cold|con|1|1|1d6|
irresistibledance|Неудержимая пляска Отто|6|Очарование|bard,wizard|Цель пляшет без контроля.
planarally|Союзник из иного плана|6|Вызов|cleric|Небожитель служит за плату.
programmedillusion|Запрограммированная иллюзия|6|Иллюзия|bard,wizard|Иллюзия по триггеру.
sunbeam|Солнечный луч|6|Воплощение|cleric,druid,sorcerer,wizard|Линия: 6d8 излучения и слепота.|save|6d8|radiant|con|1|1||
transportviaplants|Перемещение через растения|6|Вызов|druid|Телепорт по растениям.
trueseeing|Истинное зрение|6|Прорицание|bard,cleric,sorcerer,wizard|Видите истину на 120 футов.
wallofice|Ледяная стена|6|Воплощение|wizard|Стена льда: 10d6 холода.
wallofthorns|Стена шипов|6|Вызов|druid|Стена шипов: 7d8 колющего.|save|7d8|piercing|dex|1|1|1d8|
wordofrecall|Слово возвращения|6|Вызов|cleric|Телепорт в святилище.
arcanegate|Арканные врата|6|Вызов|sorcerer,warlock,wizard|Портал между двумя точками.
mentalprison|Ментальная тюрьма|6|Иллюзия|sorcerer,warlock,wizard|Иллюзия 10d10 психического.|save|5d10|psychic|int||||
conjurecelestial|Призыв небожителя|7|Вызов|cleric|Небесное существо служит вам.
delayedblastfireball|Замедленный огненный шар|7|Воплощение|sorcerer,wizard|Огненный шар растёт и взрывается.|save|12d6|fire|dex|1|1|1d6|
divineword|Слово бога|7|Воплощение|cleric|Слово оглушает, ослепляет, убивает.
etherealness|Эфирность|7|Преобразование|bard,cleric,warlock,sorcerer,wizard|Уходите в Эфирный план.
fingerofdeath|Палец смерти|7|Некромантия|sorcerer,warlock,wizard|7d8+30 некротического.|save|7d8+30|necrotic|con|1|||
firestorm|Огненный шторм|7|Воплощение|cleric,druid,sorcerer|7d10 огня по области.|save|7d10|fire|dex|1|1||
forcecage|Силовая клетка|7|Воплощение|bard,warlock,wizard|Непробиваемая клетка.
miragearcane|Мираж|7|Иллюзия|bard,druid,wizard|Меняет вид местности.
magnificentmansion|Великолепный особняк Мордекайнена|7|Вызов|bard,wizard|Особняк на день.
mordsword|Меч Мордекайнена|7|Воплощение|bard,wizard|Мечущий меч: 3d10 силой.|attack|3d10|force|||||
planeshift|Смена плана|7|Вызов|cleric,druid,sorcerer,warlock,wizard|Перенос на другой план.
prismaticspray|Радужные брызги|7|Воплощение|bard,sorcerer,wizard|Конус цветных лучей.
projectimage|Проекция образа|7|Иллюзия|bard,wizard|Двойник на расстоянии.
regenerate|Регенерация|7|Преобразование|bard,cleric,druid|Восстанавливает члены.|heal|4d8+15||||||
resurrection|Воскрешение|7|Некромантия|bard,cleric,paladin|Возвращает мёртвого до 100 лет.
reversegravity|Обратная гравитация|7|Преобразование|druid,sorcerer,wizard|Все падают вверх.
sequester|Скрытие|7|Преобразование|wizard|Скрывает предмет или существо.
simulacrum|Двойник|7|Иллюзия|wizard|Снежная копия с половиной сил.
symbol|Символ|7|Ограждение|bard,cleric,druid,wizard|Охранный символ с эффектом.
teleport|Телепортация|7|Вызов|bard,sorcerer,wizard|Мгновенный перенос.
crownofstars|Корона звёзд|7|Воплощение|sorcerer,warlock,wizard|7 звёзд бьют по 4d12.
powerwordpain|Слово силы: боль|7|Очарование|sorcerer,warlock,wizard|Парализующая боль.
whirlwind|Вихрь|7|Воплощение|druid|Вихрь: 10d6 дробящего.|save|10d6|bludgeoning|dex|1|1||
draconictransformation|Драконье превращение|7|Преобразование|sorcerer,druid,wizard|Крылья, дыхание и взгляд.
animalshapes|Звериные облики|8|Преобразование|druid,ranger|Группа превращается в зверей.
antimagicfield|Антимагическое поле|8|Ограждение|cleric,wizard|Сфера подавления магии.
antipathy|Антипатия/симпатия|8|Очарование|bard,druid,wizard|Притягивает или отталкивает.
clone|Клонирование|8|Некромантия|wizard|Запасное тело.
controlweather|Власть над погодой|8|Преобразование|cleric,druid,wizard|Меняет погоду.
demiplane|Малый план|8|Вызов|warlock,wizard|Создаёт комнату в иной реальности.
dominatemonster|Подчинение монстра|8|Очарование|bard,sorcerer,warlock,wizard|Подчиняет любое существо.
earthquake|Землетрясение|8|Воплощение|cleric,druid,sorcerer|Сотрясает область.
feeblemind|Слабоумие|8|Очарование|bard,druid,warlock,wizard|4d6 психического и ум 1.|save|4d6|psychic|int||||
glibness|Красноречие|8|Преобразование|bard,warlock|Лжёте безупречно.
holyaura|Священная аура|8|Ограждение|cleric|Союзники защищены от зла.
incendiarycloud|Огненное облако|8|Вызов|sorcerer,wizard|10d8 огня.|save|10d8|fire|dex|1|1||
maze|Лабиринт|8|Вызов|wizard|Цель заключена в лабиринте.
mindblank|Пустой разум|8|Ограждение|bard,wizard|Иммунитет к чтению разума.
powerwordstun|Слово силы: оглушение|8|Очарование|bard,sorcerer,warlock,wizard|Оглушает слабую цель.
sunburst|Солнечный взрыв|8|Воплощение|druid,sorcerer,wizard|12d6 излучения и слепота.|save|12d6|radiant|con||1||
telepathy|Телепатия|8|Воплощение|wizard|Мысленная связь на любом расстоянии.
tsunami|Цунами|8|Вызов|druid|Огромная волна: 6d10.|save|6d10|bludgeoning|str|1|1||
illusorydragon|Иллюзорный дракон|8|Иллюзия|wizard|Страшный образ дракона.
horridwilting|Ужасное иссушение|8|Некромантия|sorcerer,wizard|12d8 некротического.|save|12d8|necrotic|con|1|1||
astralprojection|Астральная проекция|9|Некромантия|cleric,warlock,wizard|Путешествие по планам.
foresight|Предвидение|9|Прорицание|bard,druid,warlock,wizard|Преимущество на всё на 8 часов.
gate|Врата|9|Вызов|cleric,sorcerer,warlock,wizard|Открывает портал на иной план.
imprisonment|Заточение|9|Ограждение|warlock,wizard|Заключает цель надолго.
massheal|Массовое исцеление|9|Воплощение|cleric|Лечит до 700 HP.|heal|100d12||||1||
meteorswarm|Метеоритный дождь|9|Воплощение|sorcerer,wizard|Четыре метеора по 20d6 огня и 20d6 дробящего.|save|20d6|fire|dex|1|1||
powerwordheal|Слово силы: исцеление|9|Воплощение|bard|Полное исцеление.|heal|100d12||||||
powerwordkill|Слово силы: смерть|9|Очарование|bard,sorcerer,warlock,wizard|Убивает цель с 100 HP или меньше.
prismaticwall|Радужная стена|9|Ограждение|wizard|Стена цветных слоёв.
shapechange|Смена формы|9|Преобразование|druid,wizard|Превращение в любое существо.
stormofvengeance|Буря возмездия|9|Вызов|druid|Адский шторм.
timestop|Остановка времени|9|Преобразование|sorcerer,wizard|Несколько ходов подряд.
truepolymorph|Истинное превращение|9|Преобразование|bard,wizard|Любое превращение навсегда.
weird|Жуть|9|Иллюзия|wizard|Самый страшный кошмар.|save|4d10|psychic|wis||1||
wish|Желание|9|Вызов|sorcerer,wizard|Самое мощное заклинание.
invulnerability|Неуязвимость|9|Ограждение|wizard|Иммунитет ко всему урону.
psychicscream|Психический вопль|9|Очарование|bard,sorcerer,warlock,wizard|14d6 психического, оглушение.|save|14d6|psychic|int|1|1||
masspolymorph|Массовое превращение|9|Преобразование|bard,sorcerer,wizard|Превращает многих.
`
