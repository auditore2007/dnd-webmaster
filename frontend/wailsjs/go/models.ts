export namespace bestiary {
	
	export class Monster {
	    id: string;
	    name: string;
	    cr: string;
	    kind: string;
	    ac: number;
	    hp: number;
	    attack: number;
	    speed: number;
	    abilities: number[];
	    weapon: model.Weapon;
	    weapons: model.Weapon[];
	    multi: number[];
	    traits: string[];
	    resist: string[];
	    vuln: string[];
	    immune: string[];
	    specials: model.Special[];
	    legendary: number;
	    legRes: number;
	    legActs: model.LegAct[];
	    lair: model.Special[];
	    phase?: model.Phase;
	    custom: boolean;
	    desc: string;
	
	    static createFrom(source: any = {}) {
	        return new Monster(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.cr = source["cr"];
	        this.kind = source["kind"];
	        this.ac = source["ac"];
	        this.hp = source["hp"];
	        this.attack = source["attack"];
	        this.speed = source["speed"];
	        this.abilities = source["abilities"];
	        this.weapon = this.convertValues(source["weapon"], model.Weapon);
	        this.weapons = this.convertValues(source["weapons"], model.Weapon);
	        this.multi = source["multi"];
	        this.traits = source["traits"];
	        this.resist = source["resist"];
	        this.vuln = source["vuln"];
	        this.immune = source["immune"];
	        this.specials = this.convertValues(source["specials"], model.Special);
	        this.legendary = source["legendary"];
	        this.legRes = source["legRes"];
	        this.legActs = this.convertValues(source["legActs"], model.LegAct);
	        this.lair = this.convertValues(source["lair"], model.Special);
	        this.phase = this.convertValues(source["phase"], model.Phase);
	        this.custom = source["custom"];
	        this.desc = source["desc"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace combat {
	
	export class Encounter {
	    order: string[];
	    init: Record<string, number>;
	    turn: number;
	    round: number;
	    acted: boolean;
	    left: number;
	    log: string[];
	    seq: number;
	    won: string;
	    react: Record<string, boolean>;
	    auto: boolean;
	    morale: boolean;
	    lair: boolean;
	    lairRound: number;
	    gone: string[];
	
	    static createFrom(source: any = {}) {
	        return new Encounter(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.order = source["order"];
	        this.init = source["init"];
	        this.turn = source["turn"];
	        this.round = source["round"];
	        this.acted = source["acted"];
	        this.left = source["left"];
	        this.log = source["log"];
	        this.seq = source["seq"];
	        this.won = source["won"];
	        this.react = source["react"];
	        this.auto = source["auto"];
	        this.morale = source["morale"];
	        this.lair = source["lair"];
	        this.lairRound = source["lairRound"];
	        this.gone = source["gone"];
	    }
	}

}

export namespace main {
	
	export class CharView {
	    id: string;
	    name: string;
	    kind: string;
	    ruleset: string;
	    race: string;
	    subrace: string;
	    class: string;
	    level: number;
	    xp: number;
	    points: number;
	    abilities: Record<string, number>;
	    hp: number;
	    tempHp: number;
	    mp: number;
	    acBonus: number;
	    conditions: string[];
	    weapons: model.Weapon[];
	    inventory: model.Item[];
	    spells: model.Spell[];
	    slotsUsed: number[];
	    deathOk: number;
	    deathFail: number;
	    stable: boolean;
	    dead: boolean;
	    active: Record<string, boolean>;
	    used: Record<string, boolean>;
	    stat?: model.MonsterStat;
	    subclass: string;
	    skills: string[];
	    expert: string[];
	    portrait: string;
	    counters: Record<string, number>;
	    notes: string;
	    effects: model.Effect[];
	    conc: string;
	    concDmg: number;
	    purse: Record<string, number>;
	    derived: model.Derived;
	    msg?: string;
	
	    static createFrom(source: any = {}) {
	        return new CharView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.kind = source["kind"];
	        this.ruleset = source["ruleset"];
	        this.race = source["race"];
	        this.subrace = source["subrace"];
	        this.class = source["class"];
	        this.level = source["level"];
	        this.xp = source["xp"];
	        this.points = source["points"];
	        this.abilities = source["abilities"];
	        this.hp = source["hp"];
	        this.tempHp = source["tempHp"];
	        this.mp = source["mp"];
	        this.acBonus = source["acBonus"];
	        this.conditions = source["conditions"];
	        this.weapons = this.convertValues(source["weapons"], model.Weapon);
	        this.inventory = this.convertValues(source["inventory"], model.Item);
	        this.spells = this.convertValues(source["spells"], model.Spell);
	        this.slotsUsed = source["slotsUsed"];
	        this.deathOk = source["deathOk"];
	        this.deathFail = source["deathFail"];
	        this.stable = source["stable"];
	        this.dead = source["dead"];
	        this.active = source["active"];
	        this.used = source["used"];
	        this.stat = this.convertValues(source["stat"], model.MonsterStat);
	        this.subclass = source["subclass"];
	        this.skills = source["skills"];
	        this.expert = source["expert"];
	        this.portrait = source["portrait"];
	        this.counters = source["counters"];
	        this.notes = source["notes"];
	        this.effects = this.convertValues(source["effects"], model.Effect);
	        this.conc = source["conc"];
	        this.concDmg = source["concDmg"];
	        this.purse = source["purse"];
	        this.derived = this.convertValues(source["derived"], model.Derived);
	        this.msg = source["msg"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class EncounterView {
	    encounter?: combat.Encounter;
	    result?: model.AttackResult;
	
	    static createFrom(source: any = {}) {
	        return new EncounterView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.encounter = this.convertValues(source["encounter"], combat.Encounter);
	        this.result = this.convertValues(source["result"], model.AttackResult);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class LegendaryView {
	    encounter?: combat.Encounter;
	    result?: model.AttackResult;
	
	    static createFrom(source: any = {}) {
	        return new LegendaryView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.encounter = this.convertValues(source["encounter"], combat.Encounter);
	        this.result = this.convertValues(source["result"], model.AttackResult);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class MapInfo {
	    id: string;
	    name: string;
	    pins: store.Pin[];
	    grid: number;
	    feet: number;
	    hasFog: boolean;
	    places: model.Place[];
	
	    static createFrom(source: any = {}) {
	        return new MapInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.pins = this.convertValues(source["pins"], store.Pin);
	        this.grid = source["grid"];
	        this.feet = source["feet"];
	        this.hasFog = source["hasFog"];
	        this.places = this.convertValues(source["places"], model.Place);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class MonsterTurnView {
	    encounter?: combat.Encounter;
	    actor: string;
	    results: model.AttackResult[];
	    targets: string[];
	
	    static createFrom(source: any = {}) {
	        return new MonsterTurnView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.encounter = this.convertValues(source["encounter"], combat.Encounter);
	        this.actor = source["actor"];
	        this.results = this.convertValues(source["results"], model.AttackResult);
	        this.targets = source["targets"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class MultiView {
	    encounter?: combat.Encounter;
	    results: model.AttackResult[];
	
	    static createFrom(source: any = {}) {
	        return new MultiView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.encounter = this.convertValues(source["encounter"], combat.Encounter);
	        this.results = this.convertValues(source["results"], model.AttackResult);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class RollView {
	    label: string;
	    rolls: number[];
	    kept: number;
	    bonus: number;
	    total: number;
	    dc: number;
	    success: boolean;
	    sides: number;
	    mode: string;
	    kind: string;
	
	    static createFrom(source: any = {}) {
	        return new RollView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.label = source["label"];
	        this.rolls = source["rolls"];
	        this.kept = source["kept"];
	        this.bonus = source["bonus"];
	        this.total = source["total"];
	        this.dc = source["dc"];
	        this.success = source["success"];
	        this.sides = source["sides"];
	        this.mode = source["mode"];
	        this.kind = source["kind"];
	    }
	}
	export class SnapshotInfo {
	    id: string;
	    name: string;
	    time: string;
	
	    static createFrom(source: any = {}) {
	        return new SnapshotInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.time = source["time"];
	    }
	}

}

export namespace model {
	
	export class AttackResult {
	    hit: boolean;
	    crit: boolean;
	    rolls: number[];
	    kept: number;
	    total: number;
	    target: number;
	    damage: number;
	    damageText: string;
	
	    static createFrom(source: any = {}) {
	        return new AttackResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.hit = source["hit"];
	        this.crit = source["crit"];
	        this.rolls = source["rolls"];
	        this.kept = source["kept"];
	        this.total = source["total"];
	        this.target = source["target"];
	        this.damage = source["damage"];
	        this.damageText = source["damageText"];
	    }
	}
	export class Effect {
	    name: string;
	    rounds: number;
	    cond: string;
	    src: string;
	    conc: boolean;
	    save: string;
	    dc: number;
	    ac: number;
	    atk: string;
	    once: boolean;
	    id: number;
	
	    static createFrom(source: any = {}) {
	        return new Effect(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.rounds = source["rounds"];
	        this.cond = source["cond"];
	        this.src = source["src"];
	        this.conc = source["conc"];
	        this.save = source["save"];
	        this.dc = source["dc"];
	        this.ac = source["ac"];
	        this.atk = source["atk"];
	        this.once = source["once"];
	        this.id = source["id"];
	    }
	}
	export class Phase {
	    pct: number;
	    name: string;
	    atk: number;
	    ac: number;
	    extra: number;
	    temp: number;
	    recharge: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Phase(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.pct = source["pct"];
	        this.name = source["name"];
	        this.atk = source["atk"];
	        this.ac = source["ac"];
	        this.extra = source["extra"];
	        this.temp = source["temp"];
	        this.recharge = source["recharge"];
	    }
	}
	export class LegAct {
	    key: string;
	    name: string;
	    cost: number;
	    weapon: number;
	    special?: Special;
	
	    static createFrom(source: any = {}) {
	        return new LegAct(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.key = source["key"];
	        this.name = source["name"];
	        this.cost = source["cost"];
	        this.weapon = source["weapon"];
	        this.special = this.convertValues(source["special"], Special);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Special {
	    key: string;
	    name: string;
	    mode: string;
	    attack: boolean;
	    dmg: string;
	    type: string;
	    xDmg: string;
	    xType: string;
	    save: string;
	    rSave: string;
	    dc: number;
	    half: boolean;
	    cond: string;
	    rounds: number;
	    repeat: boolean;
	    recharge: boolean;
	    once: boolean;
	    desc: string;
	    onlyKind: string;
	
	    static createFrom(source: any = {}) {
	        return new Special(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.key = source["key"];
	        this.name = source["name"];
	        this.mode = source["mode"];
	        this.attack = source["attack"];
	        this.dmg = source["dmg"];
	        this.type = source["type"];
	        this.xDmg = source["xDmg"];
	        this.xType = source["xType"];
	        this.save = source["save"];
	        this.rSave = source["rSave"];
	        this.dc = source["dc"];
	        this.half = source["half"];
	        this.cond = source["cond"];
	        this.rounds = source["rounds"];
	        this.repeat = source["repeat"];
	        this.recharge = source["recharge"];
	        this.once = source["once"];
	        this.desc = source["desc"];
	        this.onlyKind = source["onlyKind"];
	    }
	}
	export class MonsterStat {
	    cr: string;
	    kind: string;
	    ac: number;
	    maxHp: number;
	    attack: number;
	    speed: number;
	    traits: string[];
	    id: string;
	    multi: number[];
	    resist: string[];
	    vuln: string[];
	    immune: string[];
	    specials: Special[];
	    legendary?: number;
	    legRes?: number;
	    legActs?: LegAct[];
	    lair?: Special[];
	    phase?: Phase;
	
	    static createFrom(source: any = {}) {
	        return new MonsterStat(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.cr = source["cr"];
	        this.kind = source["kind"];
	        this.ac = source["ac"];
	        this.maxHp = source["maxHp"];
	        this.attack = source["attack"];
	        this.speed = source["speed"];
	        this.traits = source["traits"];
	        this.id = source["id"];
	        this.multi = source["multi"];
	        this.resist = source["resist"];
	        this.vuln = source["vuln"];
	        this.immune = source["immune"];
	        this.specials = this.convertValues(source["specials"], Special);
	        this.legendary = source["legendary"];
	        this.legRes = source["legRes"];
	        this.legActs = this.convertValues(source["legActs"], LegAct);
	        this.lair = this.convertValues(source["lair"], Special);
	        this.phase = this.convertValues(source["phase"], Phase);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class SpellAuto {
	    mode: string;
	    dmg: string;
	    type: string;
	    save: string;
	    half: boolean;
	    area: boolean;
	    up: string;
	    scale: boolean;
	
	    static createFrom(source: any = {}) {
	        return new SpellAuto(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.mode = source["mode"];
	        this.dmg = source["dmg"];
	        this.type = source["type"];
	        this.save = source["save"];
	        this.half = source["half"];
	        this.area = source["area"];
	        this.up = source["up"];
	        this.scale = source["scale"];
	    }
	}
	export class Spell {
	    name: string;
	    level: number;
	    cost: number;
	    note: string;
	    ref: string;
	    auto?: SpellAuto;
	
	    static createFrom(source: any = {}) {
	        return new Spell(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.level = source["level"];
	        this.cost = source["cost"];
	        this.note = source["note"];
	        this.ref = source["ref"];
	        this.auto = this.convertValues(source["auto"], SpellAuto);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Item {
	    id: string;
	    name: string;
	    qty: number;
	    weight: number;
	    desc: string;
	    acBonus: number;
	    ability: string;
	    abilityBonus: number;
	    weapon?: Weapon;
	    equipped: boolean;
	    armorBase: number;
	    armorType: string;
	    cat: string;
	    price: string;
	    rarity: string;
	
	    static createFrom(source: any = {}) {
	        return new Item(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.qty = source["qty"];
	        this.weight = source["weight"];
	        this.desc = source["desc"];
	        this.acBonus = source["acBonus"];
	        this.ability = source["ability"];
	        this.abilityBonus = source["abilityBonus"];
	        this.weapon = this.convertValues(source["weapon"], Weapon);
	        this.equipped = source["equipped"];
	        this.armorBase = source["armorBase"];
	        this.armorType = source["armorType"];
	        this.cat = source["cat"];
	        this.price = source["price"];
	        this.rarity = source["rarity"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Weapon {
	    name: string;
	    dice: string;
	    ability: string;
	    type: string;
	    finesse: boolean;
	    ranged: boolean;
	    bonus: number;
	
	    static createFrom(source: any = {}) {
	        return new Weapon(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.dice = source["dice"];
	        this.ability = source["ability"];
	        this.type = source["type"];
	        this.finesse = source["finesse"];
	        this.ranged = source["ranged"];
	        this.bonus = source["bonus"];
	    }
	}
	export class Character {
	    id: string;
	    name: string;
	    kind: string;
	    ruleset: string;
	    race: string;
	    subrace: string;
	    class: string;
	    level: number;
	    xp: number;
	    points: number;
	    abilities: Record<string, number>;
	    hp: number;
	    tempHp: number;
	    mp: number;
	    acBonus: number;
	    conditions: string[];
	    weapons: Weapon[];
	    inventory: Item[];
	    spells: Spell[];
	    slotsUsed: number[];
	    deathOk: number;
	    deathFail: number;
	    stable: boolean;
	    dead: boolean;
	    active: Record<string, boolean>;
	    used: Record<string, boolean>;
	    stat?: MonsterStat;
	    subclass: string;
	    skills: string[];
	    expert: string[];
	    portrait: string;
	    counters: Record<string, number>;
	    notes: string;
	    effects: Effect[];
	    conc: string;
	    concDmg: number;
	    purse: Record<string, number>;
	
	    static createFrom(source: any = {}) {
	        return new Character(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.kind = source["kind"];
	        this.ruleset = source["ruleset"];
	        this.race = source["race"];
	        this.subrace = source["subrace"];
	        this.class = source["class"];
	        this.level = source["level"];
	        this.xp = source["xp"];
	        this.points = source["points"];
	        this.abilities = source["abilities"];
	        this.hp = source["hp"];
	        this.tempHp = source["tempHp"];
	        this.mp = source["mp"];
	        this.acBonus = source["acBonus"];
	        this.conditions = source["conditions"];
	        this.weapons = this.convertValues(source["weapons"], Weapon);
	        this.inventory = this.convertValues(source["inventory"], Item);
	        this.spells = this.convertValues(source["spells"], Spell);
	        this.slotsUsed = source["slotsUsed"];
	        this.deathOk = source["deathOk"];
	        this.deathFail = source["deathFail"];
	        this.stable = source["stable"];
	        this.dead = source["dead"];
	        this.active = source["active"];
	        this.used = source["used"];
	        this.stat = this.convertValues(source["stat"], MonsterStat);
	        this.subclass = source["subclass"];
	        this.skills = source["skills"];
	        this.expert = source["expert"];
	        this.portrait = source["portrait"];
	        this.counters = source["counters"];
	        this.notes = source["notes"];
	        this.effects = this.convertValues(source["effects"], Effect);
	        this.conc = source["conc"];
	        this.concDmg = source["concDmg"];
	        this.purse = source["purse"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class SlotView {
	    level: number;
	    max: number;
	
	    static createFrom(source: any = {}) {
	        return new SlotView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.level = source["level"];
	        this.max = source["max"];
	    }
	}
	export class PoolView {
	    key: string;
	    name: string;
	    left: number;
	    max: number;
	    rest: string;
	
	    static createFrom(source: any = {}) {
	        return new PoolView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.key = source["key"];
	        this.name = source["name"];
	        this.left = source["left"];
	        this.max = source["max"];
	        this.rest = source["rest"];
	    }
	}
	export class Feature {
	    key: string;
	    name: string;
	    kind: string;
	    on: boolean;
	    desc: string;
	    mode: string;
	    target: string;
	    pool: string;
	    cost: number;
	    left: number;
	    max: number;
	    rest: string;
	
	    static createFrom(source: any = {}) {
	        return new Feature(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.key = source["key"];
	        this.name = source["name"];
	        this.kind = source["kind"];
	        this.on = source["on"];
	        this.desc = source["desc"];
	        this.mode = source["mode"];
	        this.target = source["target"];
	        this.pool = source["pool"];
	        this.cost = source["cost"];
	        this.left = source["left"];
	        this.max = source["max"];
	        this.rest = source["rest"];
	    }
	}
	export class Derived {
	    maxHp: number;
	    maxMp: number;
	    ac: number;
	    prof: number;
	    atkBonus: number;
	    initiative: number;
	    speed: number;
	    mods: Record<string, number>;
	    saves: Record<string, number>;
	    traits: string[];
	    effects: string[];
	    features: Feature[];
	    pools: PoolView[];
	    weapons: Weapon[];
	    slots: SlotView[];
	    load: number;
	    carry: number;
	    nextXp: number;
	    skills: Record<string, number>;
	    attacks: number;
	    spellDc: number;
	    spellAtk: number;
	
	    static createFrom(source: any = {}) {
	        return new Derived(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.maxHp = source["maxHp"];
	        this.maxMp = source["maxMp"];
	        this.ac = source["ac"];
	        this.prof = source["prof"];
	        this.atkBonus = source["atkBonus"];
	        this.initiative = source["initiative"];
	        this.speed = source["speed"];
	        this.mods = source["mods"];
	        this.saves = source["saves"];
	        this.traits = source["traits"];
	        this.effects = source["effects"];
	        this.features = this.convertValues(source["features"], Feature);
	        this.pools = this.convertValues(source["pools"], PoolView);
	        this.weapons = this.convertValues(source["weapons"], Weapon);
	        this.slots = this.convertValues(source["slots"], SlotView);
	        this.load = source["load"];
	        this.carry = source["carry"];
	        this.nextXp = source["nextXp"];
	        this.skills = source["skills"];
	        this.attacks = source["attacks"];
	        this.spellDc = source["spellDc"];
	        this.spellAtk = source["spellAtk"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	
	export class Foe {
	    id: string;
	    name: string;
	    cr: string;
	    count: number;
	
	    static createFrom(source: any = {}) {
	        return new Foe(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.cr = source["cr"];
	        this.count = source["count"];
	    }
	}
	
	
	
	export class NPC {
	    name: string;
	    race: string;
	    role: string;
	    trait: string;
	    want: string;
	    secret: string;
	    attitude: string;
	
	    static createFrom(source: any = {}) {
	        return new NPC(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.race = source["race"];
	        this.role = source["role"];
	        this.trait = source["trait"];
	        this.want = source["want"];
	        this.secret = source["secret"];
	        this.attitude = source["attitude"];
	    }
	}
	
	export class Quest {
	    title: string;
	    text: string;
	    giver: string;
	    target: string;
	    reward: string;
	    xp: number;
	
	    static createFrom(source: any = {}) {
	        return new Quest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.title = source["title"];
	        this.text = source["text"];
	        this.giver = source["giver"];
	        this.target = source["target"];
	        this.reward = source["reward"];
	        this.xp = source["xp"];
	    }
	}
	export class Place {
	    id: string;
	    x: number;
	    y: number;
	    kind: string;
	    name: string;
	    race: string;
	    note: string;
	    npcs: NPC[];
	    quests: Quest[];
	    foes: Foe[];
	    loot: string;
	
	    static createFrom(source: any = {}) {
	        return new Place(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.x = source["x"];
	        this.y = source["y"];
	        this.kind = source["kind"];
	        this.name = source["name"];
	        this.race = source["race"];
	        this.note = source["note"];
	        this.npcs = this.convertValues(source["npcs"], NPC);
	        this.quests = this.convertValues(source["quests"], Quest);
	        this.foes = this.convertValues(source["foes"], Foe);
	        this.loot = source["loot"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	
	
	
	
	
	export class SpellTemplate {
	    id: string;
	    name: string;
	    level: number;
	    school: string;
	    desc: string;
	    auto?: SpellAuto;
	
	    static createFrom(source: any = {}) {
	        return new SpellTemplate(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.level = source["level"];
	        this.school = source["school"];
	        this.desc = source["desc"];
	        this.auto = this.convertValues(source["auto"], SpellAuto);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace rules {
	
	export class Ability {
	    id: string;
	    name: string;
	
	    static createFrom(source: any = {}) {
	        return new Ability(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	    }
	}
	export class SpellDef {
	    id: string;
	    name: string;
	    level: number;
	    school: string;
	    classes: string[];
	    desc: string;
	    mode: string;
	    area: boolean;
	    conc: boolean;
	    targets: string;
	
	    static createFrom(source: any = {}) {
	        return new SpellDef(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.level = source["level"];
	        this.school = source["school"];
	        this.classes = source["classes"];
	        this.desc = source["desc"];
	        this.mode = source["mode"];
	        this.area = source["area"];
	        this.conc = source["conc"];
	        this.targets = source["targets"];
	    }
	}
	export class Skill {
	    id: string;
	    name: string;
	    ability: string;
	
	    static createFrom(source: any = {}) {
	        return new Skill(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.ability = source["ability"];
	    }
	}
	export class Subclass {
	    id: string;
	    name: string;
	    features: Feat[];
	
	    static createFrom(source: any = {}) {
	        return new Subclass(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.features = this.convertValues(source["features"], Feat);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Feat {
	    level: number;
	    name: string;
	    desc: string;
	
	    static createFrom(source: any = {}) {
	        return new Feat(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.level = source["level"];
	        this.name = source["name"];
	        this.desc = source["desc"];
	    }
	}
	export class Class {
	    id: string;
	    name: string;
	    hitDie: number;
	    saves: string[];
	    features: Feat[];
	    subclasses: Subclass[];
	    desc: string;
	    lore: string;
	
	    static createFrom(source: any = {}) {
	        return new Class(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.hitDie = source["hitDie"];
	        this.saves = source["saves"];
	        this.features = this.convertValues(source["features"], Feat);
	        this.subclasses = this.convertValues(source["subclasses"], Subclass);
	        this.desc = source["desc"];
	        this.lore = source["lore"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Race {
	    id: string;
	    name: string;
	    speed: number;
	    bonus: Record<string, number>;
	    traits: string[];
	    resist: string[];
	    subs: Race[];
	    desc: string;
	    lore: string;
	
	    static createFrom(source: any = {}) {
	        return new Race(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.speed = source["speed"];
	        this.bonus = source["bonus"];
	        this.traits = source["traits"];
	        this.resist = source["resist"];
	        this.subs = this.convertValues(source["subs"], Race);
	        this.desc = source["desc"];
	        this.lore = source["lore"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Catalog {
	    id: string;
	    name: string;
	    acLabel: string;
	    deathSaves: boolean;
	    abilities: Ability[];
	    races: Race[];
	    classes: Class[];
	    conditions: string[];
	    damageTypes: Ability[];
	    rests: Ability[];
	    skills: Skill[];
	    spells: SpellDef[];
	
	    static createFrom(source: any = {}) {
	        return new Catalog(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.acLabel = source["acLabel"];
	        this.deathSaves = source["deathSaves"];
	        this.abilities = this.convertValues(source["abilities"], Ability);
	        this.races = this.convertValues(source["races"], Race);
	        this.classes = this.convertValues(source["classes"], Class);
	        this.conditions = source["conditions"];
	        this.damageTypes = this.convertValues(source["damageTypes"], Ability);
	        this.rests = this.convertValues(source["rests"], Ability);
	        this.skills = this.convertValues(source["skills"], Skill);
	        this.spells = this.convertValues(source["spells"], SpellDef);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	
	
	
	

}

export namespace store {
	
	export class Pin {
	    x: number;
	    y: number;
	    text: string;
	
	    static createFrom(source: any = {}) {
	        return new Pin(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.x = source["x"];
	        this.y = source["y"];
	        this.text = source["text"];
	    }
	}

}

export namespace treasure {
	
	export class Treasure {
	    cr: string;
	    hoard: boolean;
	    coins: Record<string, number>;
	    items: model.Item[];
	    text: string;
	
	    static createFrom(source: any = {}) {
	        return new Treasure(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.cr = source["cr"];
	        this.hoard = source["hoard"];
	        this.coins = source["coins"];
	        this.items = this.convertValues(source["items"], model.Item);
	        this.text = source["text"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

