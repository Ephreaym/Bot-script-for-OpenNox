package BotWars

import (
	"image/color"

	"github.com/noxworld-dev/noxscript/ns/v4"
	"github.com/noxworld-dev/noxscript/ns/v4/audio"
	"github.com/noxworld-dev/noxscript/ns/v4/enchant"
	"github.com/noxworld-dev/noxscript/ns/v4/spell"
	"github.com/noxworld-dev/noxscript/ns/v4/subclass"
	"github.com/noxworld-dev/opennox-lib/object"
	"github.com/noxworld-dev/opennox-lib/script"
)

// NewConjurer creates a new Conjurer bot.
func NewConjurer(t *Team) *Conjurer {
	con := &Conjurer{team: t}
	con.init()
	return con
}

// Conjurer bot class.
type Conjurer struct {
	team              *Team
	unit              ns.Obj
	cursor            ns.Pointf
	target            ns.Obj
	bomber1           ns.Obj
	bomber2           ns.Obj
	mana              int
	startingEquipment struct {
		StreetSneakers ns.Obj
		StreetPants    ns.Obj
		StreetShirt    ns.Obj
	}
	spells struct {
		isAlive              bool
		Ready                bool // Duration unknown.
		InfravisionReady     bool // Duration is 30 seconds.
		VampirismReady       bool // Duration is 30 seconds.
		CounterspellReady    bool
		BlinkReady           bool // No real cooldown, "cooldown" implemented for balance reasons. TODO: Make random.
		FistOfVengeanceReady bool // No real cooldown, mana cost 60.
		StunReady            bool // No real cooldown.
		SummonBomber1Ready   bool // No real cooldown.
		SummonBomber2Ready   bool
		SummonGhostReady     bool
		ProtFromFireReady    bool // Duration is 60 seconds.
		ProtFromPoisonReady  bool
		ProtFromShockReady   bool
		PixieSwarmReady      bool
		ForceOfNatureReady   bool
		InversionReady       bool
		ToxicCloudReady      bool // 60 mana.
		SlowReady            bool
		MeteorReady          bool
		LesserHealReady      bool
		BurnReady            bool
	}
	summons struct {
		CreatureCage  int
		ExcludeSummon []ns.Obj
		BomberCount   int
	}
	audio struct {
		ManaRestoreSound bool
	}
	behaviour struct {
		AntiStuck bool
		Busy      bool
	}
	reactionTime int
}

func (con *Conjurer) init() {
	// Reset spells ConBot.
	con.spells.Ready = true
	// Debuff spells.
	con.spells.SlowReady = true
	con.spells.StunReady = true
	// Offensive spells.
	con.spells.MeteorReady = true
	con.spells.BurnReady = true
	con.spells.FistOfVengeanceReady = true
	con.spells.PixieSwarmReady = true
	con.spells.ForceOfNatureReady = true
	con.spells.ToxicCloudReady = true
	// Defensive spells.
	con.spells.BlinkReady = true
	con.spells.CounterspellReady = true
	con.spells.InversionReady = true
	con.spells.LesserHealReady = true
	// Summons.
	con.summons.CreatureCage = 0
	con.summons.BomberCount = 0
	con.spells.SummonGhostReady = true
	con.spells.SummonBomber1Ready = true
	con.spells.SummonBomber2Ready = true
	// Buff spells.
	con.spells.InfravisionReady = true
	con.spells.VampirismReady = true
	con.spells.ProtFromFireReady = true
	con.spells.ProtFromPoisonReady = true
	con.spells.ProtFromShockReady = true
	// Behaviour.
	con.behaviour.AntiStuck = true
	con.behaviour.Busy = false
	// Create ConBot.
	con.unit = ns.CreateObject("NPC", con.team.SpawnPoint())
	con.unit.Enchant(enchant.INVULNERABLE, script.Frames(150))
	con.unit.SetMaxHealth(100)
	con.unit.SetStrength(55)
	con.unit.SetBaseSpeed(88)
	con.spells.isAlive = true
	con.mana = 125
	// Set Team.
	if GameModeIsCTF {
		con.unit.SetOwner(con.team.Spawns()[0])
	}
	con.unit.SetTeam(con.team.Team())
	if con.unit.HasTeam(ns.Teams()[0]) {
		con.unit.SetColor(0, color.NRGBA{R: 255, G: 0, B: 0, A: 255})
		con.unit.SetColor(1, color.NRGBA{R: 255, G: 0, B: 0, A: 255})
		con.unit.SetColor(2, color.NRGBA{R: 255, G: 0, B: 0, A: 255})
		con.unit.SetColor(3, color.NRGBA{R: 255, G: 0, B: 0, A: 255})
		con.unit.SetColor(4, color.NRGBA{R: 255, G: 0, B: 0, A: 255})
		con.unit.SetColor(5, color.NRGBA{R: 255, G: 0, B: 0, A: 255})
	} else {
		con.unit.SetColor(0, color.NRGBA{R: 0, G: 0, B: 255, A: 255})
		con.unit.SetColor(1, color.NRGBA{R: 0, G: 0, B: 255, A: 255})
		con.unit.SetColor(2, color.NRGBA{R: 0, G: 0, B: 255, A: 255})
		con.unit.SetColor(3, color.NRGBA{R: 0, G: 0, B: 255, A: 255})
		con.unit.SetColor(4, color.NRGBA{R: 0, G: 0, B: 255, A: 255})
		con.unit.SetColor(5, color.NRGBA{R: 0, G: 0, B: 255, A: 255})
	}
	// Create ConBot mouse cursor.
	con.target = NoTarget
	con.cursor = NoTarget.Pos()
	// Set difficulty (0 = Botlike, 15 = hard, 30 = normal, 45 = easy, 60 = beginner)
	con.reactionTime = BotDifficulty
	// Set ConBot properties.
	con.unit.MonsterStatusEnable(object.MonStatusAlwaysRun)
	con.unit.MonsterStatusEnable(object.MonStatusCanCastSpells)
	con.unit.MonsterStatusEnable(object.MonStatusAlert)
	con.unit.AggressionLevel(0.16)
	ns.NewTimer(ns.Seconds(4), func() {
		con.unit.AggressionLevel(0.83)
	})
	con.unit.Hunt()
	con.unit.ResumeLevel(0.8)
	con.unit.RetreatLevel(0.4)
	// Create and equip ConBot starting equipment. TODO: Change location of item creation OR stop them from respawning automatically.
	con.startingEquipment.StreetSneakers = ns.CreateObject("StreetSneakers", ns.Ptf(150, 150))
	con.startingEquipment.StreetPants = ns.CreateObject("StreetPants", ns.Ptf(150, 150))
	con.startingEquipment.StreetShirt = ns.CreateObject("StreetShirt", ns.Ptf(150, 150))
	con.unit.Equip(con.startingEquipment.StreetPants)
	con.unit.Equip(con.startingEquipment.StreetShirt)
	con.unit.Equip(con.startingEquipment.StreetSneakers)
	// Buff on respawn.
	con.buffInitial()
	// Enemy sighted.
	con.unit.OnEvent(ns.EventEnemySighted, con.onEnemySighted)
	// On Collision.
	con.unit.OnEvent(ns.EventCollision, con.onCollide)
	// Retreat.
	con.unit.OnEvent(ns.EventRetreat, con.onRetreat)
	// Enemy lost.
	con.unit.OnEvent(ns.EventLostEnemy, con.onLostEnemy)
	// On death.
	con.unit.OnEvent(ns.EventDeath, con.onDeath)
	// On heard.
	con.unit.OnEvent(ns.EventEnemyHeard, con.onEnemyHeard)
	con.unit.OnEvent(ns.EventIsHit, con.onHit)
	// Looking for enemies.
	con.unit.OnEvent(ns.EventLookingForEnemy, con.onLookingForTarget)
	//con.unit.OnEvent(ns.EventChangeFocus, con.onChangeFocus)
	con.unit.OnEvent(ns.EventEndOfWaypoint, con.onEndOfWaypoint)
	con.PassiveManaRegen()
	con.LookForWeapon()
	con.WeaponPreference()
	ns.OnChat(con.onConCommand)
	con.findLoot()
	//
	//con.checkSummonCount()
	//con.summontest()
}

func (con *Conjurer) summontest() {
	ns.CastSpell(spell.SUMMON_BAT, con.unit, con.unit)
	ns.NewTimer(ns.Seconds(20), func() {
		con.summontest()
	})
}

func (con *Conjurer) onHit() {
	if con.mana <= 20 && !con.behaviour.Busy {
		con.GoToManaObelisk()
	}
}

// Checks the ammount of summons active for the Conjurer bot.
func (con *Conjurer) checkSummonCount() {
	// Get all active sommons that belong to the Conjuer bot.
	allActiveSummons := ns.FindAllObjects(ns.HasClass(object.ClassMonster), ns.ObjCondFunc(func(it ns.Obj) bool {
		return it.HasOwner(con.unit)
	}))
	for _, summon := range allActiveSummons {
		if summon.HasSubclass(subclass.SMALL_MONSTER) {
			// if summon != in ExcludeSummon[]{
			// Add the summon to the Creature Cage.
			con.summons.CreatureCage = con.summons.CreatureCage + 1
			// Track if the summon is a bomber.
			if summon.HasSubclass(subclass.BOMBER) {
				con.summons.BomberCount = con.summons.BomberCount + 1
			}
			// Add summon to ExcludeSummon array to prevent it from getting counted twice.
			con.summons.ExcludeSummon = append(con.summons.ExcludeSummon, summon)
			ns.PrintStrToAll("+1 to CreatureCage")
			summon.OnEvent(ns.EventDeath, func() {
				// Remove summon from the Creature Cage on death.
				ns.PrintStrToAll("-1 to CreatureCage")
				con.summons.CreatureCage = con.summons.CreatureCage - 1
				// Track if the summon is a bomber.
				if summon.HasSubclass(subclass.BOMBER) {
					con.summons.BomberCount = con.summons.BomberCount + -1
				}
			})
			// }
		}
		if summon.HasSubclass(subclass.MEDIUM_MONSTER) {
			// if summon != in ExcludeSummon[]{
			con.summons.CreatureCage = con.summons.CreatureCage + 2
			con.summons.ExcludeSummon = append(con.summons.ExcludeSummon, summon)
			ns.PrintStrToAll("+2 to CreatureCage")
			summon.OnEvent(ns.EventDeath, func() {
				ns.PrintStrToAll("-2 to CreatureCage")
				con.summons.CreatureCage = con.summons.CreatureCage - 2
			})
			// }
		}
		if summon.HasSubclass(subclass.LARGE_MONSTER) {
			// if summon != in ExcludeSummon[]{
			con.summons.CreatureCage = con.summons.CreatureCage + 4
			con.summons.ExcludeSummon = append(con.summons.ExcludeSummon, summon)
			ns.PrintStrToAll("+4 to CreatureCage")
			summon.OnEvent(ns.EventDeath, func() {
				ns.PrintStrToAll("-4 to CreatureCage")
				con.summons.CreatureCage = con.summons.CreatureCage - 4
			})
			// }
		}
	}
	ns.NewTimer(ns.Seconds(1), func() {
		con.checkSummonCount()
	})
}

func (con *Conjurer) onEndOfWaypoint() {
	con.behaviour.Busy = false
	con.unit.AggressionLevel(0.83)
	if con.mana <= 49 {
		con.GoToManaObelisk()
	} else {
		if GameModeIsCTF {
			con.team.CheckAttackOrDefend(con.unit)
		} else {
			con.unit.WalkTo(con.target.Pos())
			ns.NewTimer(ns.Seconds(2), func() {
				con.unit.Hunt()
			})
		}
	}
	con.LookForNearbyItems()
}

func (con *Conjurer) buffInitial() {
	con.castVampirism()
}

func (con *Conjurer) onLookingForTarget() {
	con.castInfravision()
}

func (con *Conjurer) onEnemyHeard() {
	con.castForceOfNature()
}

func (con *Conjurer) onEnemySighted() {
	con.target = ns.GetCaller()
	con.castForceOfNature()
}

func (con *Conjurer) onCollide() {
	if con.spells.isAlive {
		caller := ns.GetCaller()
		if GameModeIsCTF {
			con.team.CheckPickUpEnemyFlag(caller, con.unit)
			con.team.CheckCaptureEnemyFlag(caller, con.unit)
			con.team.CheckRetrievedOwnFlag(caller, con.unit)
		}
	}
}

func (con *Conjurer) onRetreat() {
	con.castBlink()
}

func (con *Conjurer) onLostEnemy() {
	con.castInfravision()
	if GameModeIsCTF {
		con.team.WalkToOwnFlag(con.unit)
	}
}

func (con *Conjurer) onDeath() {
	con.spells.isAlive = false
	con.spells.Ready = false
	con.unit.FlagsEnable(object.FlagNoCollide)
	con.team.DropEnemyFlag(con.unit)
	con.unit.DestroyChat()
	ns.AudioEvent(audio.NPCDie, con.unit)
	// TODO: Change ns.GetHost() to correct caller. Is there no Gvar1 replacement?
	// ns.GetHost().ChangeScore(+1)
	if !GameModeIsCTF {
		if con.unit.HasTeam(ns.Teams()[0]) {
			ns.Teams()[1].ChangeScore(+1)
		} else {
			ns.Teams()[0].ChangeScore(+1)
		}
	}
	if !ItemDropEnabled {
		con.startingEquipment.StreetPants.Delete()
		con.startingEquipment.StreetShirt.Delete()
		con.startingEquipment.StreetSneakers.Delete()
	}
	ns.NewTimer(ns.Frames(60), func() {
		ns.AudioEvent(audio.BlinkCast, con.unit)
		con.unit.Delete()
		if ItemDropEnabled {
			con.startingEquipment.StreetPants.Delete()
			con.startingEquipment.StreetShirt.Delete()
			con.startingEquipment.StreetSneakers.Delete()
		}
		if BotRespawn {
			con.init()
		}
	})
}

func (con *Conjurer) PassiveManaRegen() {
	if con.spells.isAlive {
		ns.NewTimer(ns.Seconds(2), func() {
			if con.mana < 125 {
				if !BotMana {
					con.mana = con.mana + 300
				}
				con.mana = con.mana + 1
			}
			con.PassiveManaRegen()
			//ns.PrintStrToAll("con mana: " + strconv.Itoa(con.mana))
		})
	}
}

func (con *Conjurer) UsePotions() {
	if con.unit.CurrentHealth() <= 25 && con.unit.InItems().FindObjects(nil, ns.HasTypeName{"RedPotion"}) != 0 {
		ns.AudioEvent(audio.LesserHealEffect, con.unit)
		RedPotion := con.unit.Items(ns.HasTypeName{"RedPotion"})
		con.unit.SetHealth(con.unit.CurrentHealth() + 50)
		RedPotion[0].Delete()
	}
	if con.mana <= 100 && con.unit.InItems().FindObjects(nil, ns.HasTypeName{"BluePotion"}) != 0 {
		con.mana = con.mana + 50
		ns.AudioEvent(audio.RestoreMana, con.unit)
		BluePotion := con.unit.Items(ns.HasTypeName{"BluePotion"})
		BluePotion[0].Delete()
	}
}

func (con *Conjurer) GoToManaObelisk() {
	if !con.behaviour.Busy {
		con.behaviour.Busy = true
		con.unit.AggressionLevel(0.16)
		NearestObeliskWithMana := ns.FindClosestObjectIn(con.unit, ns.Objects(AllManaObelisksOnMap),
			ns.ObjCondFunc(func(it ns.Obj) bool {
				return it.CurrentMana() >= 10
			}),
		)

		if con.unit == con.team.TeamTank {
			if con.unit.CanSee(NearestObeliskWithMana) {
				con.unit.WalkTo(NearestObeliskWithMana.Pos())
			}
		} else {
			con.unit.WalkTo(NearestObeliskWithMana.Pos())
		}
	}
}

func (con *Conjurer) RestoreMana() {
	if con.mana < 125 {
		for i := 0; i < len(AllManaObelisksOnMap); i++ {
			if AllManaObelisksOnMap[i].CurrentMana() > 0 && con.unit.CanSee(AllManaObelisksOnMap[i]) && (ns.InCirclef{Center: con.unit, R: 50}).Matches(AllManaObelisksOnMap[i]) {
				con.mana = con.mana + 1
				AllManaObelisksOnMap[i].SetMana(AllManaObelisksOnMap[i].CurrentMana() - 1)
				con.RestoreManaSound()
			}
		}
	}
}

func (con *Conjurer) RestoreManaSound() {
	if !con.audio.ManaRestoreSound {
		con.audio.ManaRestoreSound = true
		ns.AudioEvent(audio.RestoreMana, con.unit)
		ns.NewTimer(ns.Frames(15), func() {
			con.audio.ManaRestoreSound = false
		})
	}
}

func (con *Conjurer) checkForMissiles() {
	// Maybe need to add a ns.hasteam condition. Not sure yet.
	if sp2 := ns.FindClosestObject(con.unit, ns.HasTypeName{"DeathBall"}, ns.InCirclef{Center: con.unit, R: 500}); sp2 != nil {
		{
			arr2 := ns.FindAllObjects(
				ns.HasTypeName{"NewPlayer", "NPC"},
				ns.HasTeam{con.team.Enemy.Team()},
			)
			for i := 0; i < len(arr2); i++ {
				if sp2.HasOwner(arr2[i]) {
					con.castCounterspellAtForceOfNature()
				}
			}
		}
	} else {
		if sp := ns.FindClosestObject(con.unit, ns.HasClass(object.ClassMissile), ns.InCirclef{Center: con.unit, R: 500}); sp != nil {
			if sp.HasOwner(con.target) {
				con.castInversion()
			}
		}
	}
}

func (con *Conjurer) Update() {
	con.checkForMissiles()
	con.UsePotions()
	con.RestoreMana()
	if con.unit.HasEnchant(enchant.ANTI_MAGIC) {
		con.spells.Ready = true
	}
	if con.unit.HasEnchant(enchant.HELD) || con.unit.HasEnchant(enchant.SLOWED) {
		con.castBlink()
	}
	if con.target.HasEnchant(enchant.HELD) || con.target.HasEnchant(enchant.SLOWED) {
		if con.unit.CanSee(con.target) {
			//con.castFistOfVengeance()
			con.castMeteor()
			con.castBurn()
			con.castToxicCloud()
		}
	}
	if con.spells.Ready && con.unit.CanSee(con.target) {
		if !GameModeIsCTF {
			con.castStun()
		}
		con.castPixieSwarm()
		con.castSlow()

	}
	if !con.unit.CanSee(con.target) && con.spells.Ready {
		con.castVampirism()
		con.castProtectionFromShock()
		con.castProtectionFromFire()
		con.castProtectionFromPoison()
		//con.summonBomber1()
		//con.summonBomber2()
	}
}

func (con *Conjurer) LookForWeapon() {
	ItemLocation := ns.FindClosestObject(con.unit, ns.HasTypeName{"CrossBow", "InfinitePainWand"})
	if ItemLocation != nil {
		con.unit.WalkTo(ItemLocation.Pos())
	}
}

func (con *Conjurer) LookForNearbyItems() {
	if ns.FindAllObjects(ns.HasTypeName{"CrossBow", "InfinitePainWand", "InfinitePainWand", "LesserFireballWand", "Quiver",
		"LeatherArmoredBoots", "LeatherArmor",
		"LeatherHelm",
		"LeatherLeggings", "LeatherArmbands",
		"RedPotion",
		"ConjurerHelm",
		"CurePoisonPotion",
		"BluePotion",
		"LeatherBoots", "MedievalCloak", "MedievalShirt", "MedievalPants"},
		ns.InCirclef{Center: con.unit, R: 200}) != nil {
		if con.unit.InItems().FindObjects(nil, ns.HasTypeName{"CrossBow", "InfinitePainWand", "InfinitePainWand", "LesserFireballWand", "Quiver",
			"LeatherArmoredBoots", "LeatherArmor",
			"LeatherHelm",
			"LeatherLeggings", "LeatherArmbands",
			"RedPotion",
			"ConjurerHelm",
			"CurePoisonPotion",
			"BluePotion",
			"LeatherBoots", "MedievalCloak", "MedievalShirt", "MedievalPants"}) == 0 {
			ItemLocation := ns.FindAllObjects(ns.HasTypeName{"CrossBow", "InfinitePainWand", "InfinitePainWand", "LesserFireballWand", "Quiver",
				"LeatherArmoredBoots", "LeatherArmor",
				"LeatherHelm",
				"LeatherLeggings", "LeatherArmbands",
				"RedPotion",
				"ConjurerHelm",
				"CurePoisonPotion",
				"BluePotion",
				"LeatherBoots", "MedievalCloak", "MedievalShirt", "MedievalPants"},
				ns.InCirclef{Center: con.unit, R: 200},
			)
			if con.unit.CanSee(ItemLocation[0]) {
				con.unit.WalkTo(ItemLocation[0].Pos())
			}
		}
	}
	ns.NewTimer(ns.Seconds(5), func() {
		// prevent bots getting stuck to stay in loop.
		if con.behaviour.AntiStuck {
			con.behaviour.AntiStuck = false
			if GameModeIsCTF {
				con.team.CheckAttackOrDefend(con.unit)
			} else {
				con.unit.Hunt()
			}
			ns.NewTimer(ns.Seconds(6), func() {
				con.behaviour.AntiStuck = true
			})
		}
	})
}

func (con *Conjurer) WeaponPreference() {
	// Priority list to get the prefered weapon.
	// TODO: Add stun and range conditions.
	if con.unit.InItems().FindObjects(nil, ns.HasTypeName{"CrossBow"}) != 0 && con.unit.InEquipment().FindObjects(nil, ns.HasTypeName{"CrossBow"}) == 0 {
		con.unit.InItems().FindObjects(
			func(it ns.Obj) bool {
				con.unit.Equip(it)
				//war.unit.Chat("I swapped to my GreatSword!")
				return true
			},
			ns.HasTypeName{"FireStormWand"},
		)
	} else if con.unit.InItems().FindObjects(nil, ns.HasTypeName{"InfinitePainWand"}) != 0 && con.unit.InEquipment().FindObjects(nil, ns.HasTypeName{"InfinitePainWand"}) == 0 {
		con.unit.InItems().FindObjects(
			func(it ns.Obj) bool {
				con.unit.Equip(it)
				//war.unit.Chat("I swapped to my WarHammer!")
				return true
			},
			ns.HasTypeName{"ForceWand"},
		)
	}
	ns.NewTimer(ns.Seconds(10), func() {
		con.WeaponPreference()
	})
}

func (con *Conjurer) findLoot() {
	const dist = 75
	// Weapons.
	weapons := ns.FindAllObjects(
		ns.InCirclef{Center: con.unit, R: dist},
		ns.HasTypeName{
			// Wands.
			"InfinitePainWand", "LesserFireballWand",
			//"SulphorousShowerWand",
			//"SulphorousFlareWand",
			//"StaffWooden",

			// Crossbow and Bow.
			"CrossBow",
			"Bow",
			"Quiver",
		},
	)
	for _, item := range weapons {
		if con.unit.CanSee(item) {
			con.unit.Pickup(item)
			con.unit.Equip(con.unit.GetLastItem())
		}
	}
	// Quiver.
	quiver := ns.FindAllObjects(
		ns.InCirclef{Center: con.unit, R: dist},
		ns.HasTypeName{
			// Quiver.
			"Quiver",
		},
	)
	for _, item := range quiver {
		if con.unit.CanSee(item) {
			con.unit.Pickup(item)
		}
	}
	// Armor.
	armor := ns.FindAllObjects(
		ns.InCirclef{Center: con.unit, R: dist},
		ns.HasTypeName{
			// BlueConjurer Helm.
			//"ConjurerHelm",

			// Leather armor.
			"LeatherArmoredBoots", "LeatherArmor",
			//"LeatherHelm",
			"LeatherLeggings", "LeatherArmbands",

			// Cloth armor.
			"LeatherBoots", "MedievalCloak", "MedievalShirt", "MedievalPants",
		},
	)
	for _, item := range armor {
		if con.unit.CanSee(item) {
			con.unit.Pickup(item)
			con.unit.Equip(con.unit.GetLastItem())
		}
	}
	// Potions.
	potions := ns.FindAllObjects(
		ns.InCirclef{Center: con.unit, R: dist},
		ns.HasTypeName{
			"RedPotion",
			"CurePoisonPotion",
			"BluePotion",
		},
	)
	for _, item := range potions {
		if con.unit.CanSee(item) {
			con.unit.Pickup(item)
		}
	}
	ns.NewTimer(ns.Frames(15), func() {
		con.findLoot()
	})
}

// ------------------------------------------------------------------------------------------------------------------------------------ //
// ---------------------------------------------------------------- SPELL BOOK -------------------------------------------------------- //
// ------------------------------------------------------------------------------------------------------------------------------------ //

func (con *Conjurer) castLesserHeal() {
	// Check if cooldowns are ready.
	if con.mana >= 30 && con.spells.isAlive && !con.unit.HasEnchant(enchant.ANTI_MAGIC) && con.unit.CurrentHealth() <= 60 && con.spells.Ready && con.spells.LesserHealReady {
		// Trigger cooldown.
		con.spells.Ready = false
		// Check reaction time based on difficulty setting.
		ns.NewTimer(ns.Frames(con.reactionTime), func() {
			// Check for War Cry before chant.
			if con.spells.isAlive && !con.unit.HasEnchant(enchant.ANTI_MAGIC) {
				castPhonemes(con.unit, []audio.Name{PhDownRight, PhUp, PhDownLeft}, func() {
					// Check for War Cry before spell release.
					if con.spells.isAlive && !con.unit.HasEnchant(enchant.ANTI_MAGIC) {
						con.spells.LesserHealReady = false
						con.mana = con.mana - 30
						ns.CastSpell(spell.LESSER_HEAL, con.unit, con.unit)
						// Global cooldown.
						ns.NewTimer(ns.Frames(3), func() {
							con.spells.Ready = true
						})
						// Shock cooldown.
						ns.NewTimer(ns.Seconds(1), func() {
							con.spells.LesserHealReady = true
						})
					}
				})
			} else {
				ns.NewTimer(ns.Frames(con.reactionTime), func() {
					con.spells.Ready = true
				})
			}
		})
	}
}

func (con *Conjurer) castVampirism() {
	// Check if cooldowns are ready.
	if con.mana >= 20 && con.spells.isAlive && !con.unit.HasEnchant(enchant.ANTI_MAGIC) && con.spells.Ready && con.spells.VampirismReady {
		// Trigger cooldown.
		con.spells.Ready = false
		// Check reaction time based on difficulty setting.
		ns.NewTimer(ns.Frames(con.reactionTime), func() {
			// Check for War Cry before chant.
			if con.spells.isAlive && !con.unit.HasEnchant(enchant.ANTI_MAGIC) {
				castPhonemes(con.unit, []audio.Name{PhUp, PhDown, PhLeft, PhRight}, func() {
					// Check for War Cry before spell release.
					if con.spells.isAlive && !con.unit.HasEnchant(enchant.ANTI_MAGIC) {
						con.spells.VampirismReady = false
						con.mana = con.mana - 20
						ns.CastSpell(spell.VAMPIRISM, con.unit, con.unit)
						// Global cooldown.
						con.spells.Ready = true
						// Vampirism cooldown.
						ns.NewTimer(ns.Seconds(30), func() {
							con.spells.VampirismReady = true
						})
					}
				})
			}
		})
	}
}

func (con *Conjurer) castProtectionFromPoison() {
	// Check if cooldowns are ready.
	if con.mana >= 30 && con.spells.isAlive && !con.unit.HasEnchant(enchant.ANTI_MAGIC) && con.spells.Ready && con.spells.ProtFromPoisonReady {
		// Trigger cooldown.
		con.spells.Ready = false
		// Check reaction time based on difficulty setting.
		ns.NewTimer(ns.Frames(con.reactionTime), func() {
			// Check for War Cry before chant.
			if con.spells.isAlive && !con.unit.HasEnchant(enchant.ANTI_MAGIC) {
				castPhonemes(con.unit, []audio.Name{PhLeft, PhRight, PhDownLeft, PhUpRight}, func() {
					// Check for War Cry before spell release.
					if con.spells.isAlive && !con.unit.HasEnchant(enchant.ANTI_MAGIC) {
						con.spells.ProtFromPoisonReady = false
						con.mana = con.mana - 30
						ns.CastSpell(spell.PROTECTION_FROM_POISON, con.unit, con.unit)
						// Global cooldown.
						con.spells.Ready = true
						// Protection From Poison cooldown.
						ns.NewTimer(ns.Seconds(60), func() {
							con.spells.ProtFromPoisonReady = true
						})
					}
				})
			}
		})
	}
}

func (con *Conjurer) castPixieSwarm() {
	// Check if cooldowns are ready.
	if con.mana >= 30 && con.spells.isAlive && !con.unit.HasEnchant(enchant.ANTI_MAGIC) && con.spells.Ready && con.spells.PixieSwarmReady {
		// Trigger cooldown.
		con.spells.Ready = false
		// Check reaction time based on difficulty setting.
		ns.NewTimer(ns.Frames(con.reactionTime), func() {
			// Check for War Cry before chant.
			if con.spells.isAlive && !con.unit.HasEnchant(enchant.ANTI_MAGIC) {
				castPhonemes(con.unit, []audio.Name{PhLeft, PhDown, PhRight, PhDown}, func() {
					// Check for War Cry before spell release.
					if con.spells.isAlive && !con.unit.HasEnchant(enchant.ANTI_MAGIC) {
						con.spells.PixieSwarmReady = false
						con.mana = con.mana - 30
						ns.CastSpell(spell.PIXIE_SWARM, con.unit, con.unit)
						// Global cooldown.
						con.spells.Ready = true
						// Pixie Swarm cooldown.
						ns.NewTimer(ns.Seconds(10), func() {
							con.spells.PixieSwarmReady = true
						})
					}
				})
			}
		})
	}
}

func (con *Conjurer) castFistOfVengeance() {
	// Check if cooldowns are ready.
	if con.mana >= 60 && con.spells.isAlive && !con.unit.HasEnchant(enchant.ANTI_MAGIC) && con.unit.CanSee(con.target) && con.spells.FistOfVengeanceReady && con.spells.Ready {
		// Select target.
		con.cursor = con.target.Pos()
		// Trigger cooldown.
		con.spells.Ready = false
		// Check reaction time based on difficulty setting.
		ns.NewTimer(ns.Frames(con.reactionTime), func() {
			// Check for War Cry before chant.
			if con.spells.isAlive && !con.unit.HasEnchant(enchant.ANTI_MAGIC) {
				castPhonemes(con.unit, []audio.Name{PhUpRight, PhUp, PhDown}, func() {
					// Check for War Cry before spell release.
					if con.spells.isAlive && !con.unit.HasEnchant(enchant.ANTI_MAGIC) {
						// Aim.
						con.unit.LookAtObject(con.target)
						con.unit.Pause(ns.Frames(con.reactionTime))
						con.spells.FistOfVengeanceReady = false
						ns.CastSpell(spell.FIST, con.unit, con.cursor)
						con.mana = con.mana - 60
						// Global cooldown.
						con.spells.Ready = true
						ns.NewTimer(ns.Seconds(10), func() {
							// Fist Of Vengeance cooldown.
							con.spells.FistOfVengeanceReady = true
						})
					}
				})
			}
		})
	}
}

func (con *Conjurer) castForceOfNature() {
	// Check if cooldowns are ready.
	if con.mana >= 60 && con.spells.isAlive && !con.unit.HasEnchant(enchant.ANTI_MAGIC) && con.spells.ForceOfNatureReady && con.spells.Ready {
		// Select target.
		// Trigger cooldown.
		con.spells.Ready = false
		// Check reaction time based on difficulty setting.
		ns.NewTimer(ns.Frames(con.reactionTime), func() {
			// Check for War Cry before chant.
			if con.spells.isAlive && !con.unit.HasEnchant(enchant.ANTI_MAGIC) {
				castPhonemes(con.unit, []audio.Name{PhDownRight, PhDownLeft, PhDown}, func() {
					// Check for War Cry before spell release.
					if con.spells.isAlive && !con.unit.HasEnchant(enchant.ANTI_MAGIC) {
						// Aim.
						con.spells.ForceOfNatureReady = false
						con.unit.LookAtObject(con.target)
						con.unit.Pause(ns.Frames(30))
						con.mana = con.mana - 60
						ns.CastSpell(spell.FORCE_OF_NATURE, con.unit, con.target)
						// Global cooldown.
						con.spells.Ready = true
						// Force of Nature cooldown.
						ns.NewTimer(ns.Seconds(20), func() {
							con.spells.ForceOfNatureReady = true
						})
					}
				})
			}
		})
	}
}

func (con *Conjurer) castProtectionFromFire() {
	// Check if cooldowns are ready.
	if con.mana >= 30 && con.spells.isAlive && !con.unit.HasEnchant(enchant.ANTI_MAGIC) && !con.unit.HasEnchant(enchant.PROTECT_FROM_FIRE) && con.spells.Ready && con.spells.ProtFromFireReady {
		// Trigger cooldown.
		con.spells.Ready = false
		// Check reaction time based on difficulty setting.
		ns.NewTimer(ns.Frames(con.reactionTime), func() {
			// Check for War Cry before chant.
			if con.spells.isAlive && !con.unit.HasEnchant(enchant.ANTI_MAGIC) {
				castPhonemes(con.unit, []audio.Name{PhLeft, PhRight, PhDownRight, PhUpLeft}, func() {
					// Check for War Cry before spell release.
					if con.spells.isAlive && !con.unit.HasEnchant(enchant.ANTI_MAGIC) {
						con.spells.ProtFromFireReady = false
						con.mana = con.mana - 30
						ns.CastSpell(spell.PROTECTION_FROM_FIRE, con.unit, con.unit)
						// Global cooldown.
						ns.NewTimer(ns.Frames(3), func() {
							con.spells.Ready = true
						})
						// Protection From Fire cooldown.
						ns.NewTimer(ns.Seconds(60), func() {
							con.spells.ProtFromFireReady = true
						})
					}
				})
			} else {
				ns.NewTimer(ns.Frames(con.reactionTime), func() {
					con.spells.Ready = true
				})
			}
		})
	}
}

func (con *Conjurer) castProtectionFromShock() {
	// Check if cooldowns are ready.
	if con.mana >= 30 && con.spells.isAlive && !con.unit.HasEnchant(enchant.ANTI_MAGIC) && !con.unit.HasEnchant(enchant.PROTECT_FROM_ELECTRICITY) && con.spells.Ready && con.spells.ProtFromShockReady {
		// Trigger cooldown.
		con.spells.Ready = false
		// Check reaction time based on difficulty setting.
		ns.NewTimer(ns.Frames(con.reactionTime), func() {
			// Check for War Cry before chant.
			if con.spells.isAlive && !con.unit.HasEnchant(enchant.ANTI_MAGIC) {
				castPhonemes(con.unit, []audio.Name{PhRight, PhLeft, PhDownRight, PhUpLeft}, func() {
					// Check for War Cry before spell release.
					if con.spells.isAlive && !con.unit.HasEnchant(enchant.ANTI_MAGIC) {
						con.spells.ProtFromShockReady = false
						con.mana = con.mana - 30
						ns.CastSpell(spell.PROTECTION_FROM_ELECTRICITY, con.unit, con.unit)
						// Global cooldown.
						ns.NewTimer(ns.Frames(3), func() {
							con.spells.Ready = true
						})
						// Protection From Shock cooldown.
						ns.NewTimer(ns.Seconds(60), func() {
							con.spells.ProtFromShockReady = true
						})
					}
				})
			} else {
				ns.NewTimer(ns.Frames(con.reactionTime), func() {
					con.spells.Ready = true
				})
			}
		})
	}
}

func (con *Conjurer) castInversion() {
	// Check if cooldowns are ready.
	if con.mana >= 10 && con.spells.isAlive && !con.unit.HasEnchant(enchant.ANTI_MAGIC) && con.spells.Ready && con.spells.InversionReady {
		// Trigger cooldown.
		con.spells.Ready = false
		// Check reaction time based on difficulty setting.
		ns.NewTimer(ns.Frames(con.reactionTime), func() {
			// Check for War Cry before chant.
			if con.spells.isAlive && !con.unit.HasEnchant(enchant.ANTI_MAGIC) {
				castPhonemes(con.unit, []audio.Name{PhUpLeft, FPhUpRight}, func() {
					// Check for War Cry before spell release.
					if con.spells.isAlive && !con.unit.HasEnchant(enchant.ANTI_MAGIC) {
						con.spells.InversionReady = false
						con.mana = con.mana - 10
						ns.CastSpell(spell.INVERSION, con.unit, con.unit)
						// Global cooldown.
						ns.NewTimer(ns.Frames(3), func() {
							con.spells.Ready = true
						})
						// Inversion cooldown.
						ns.NewTimer(ns.Seconds(1), func() {
							con.spells.InversionReady = true
						})
					}
				})
			} else {
				ns.NewTimer(ns.Frames(con.reactionTime), func() {
					con.spells.Ready = true
				})
			}
		})
	}
}

func (con *Conjurer) castBlink() {
	// Check if cooldowns are ready.
	if con.mana >= 10 && con.spells.isAlive && !con.unit.HasEnchant(enchant.ANTI_MAGIC) && con.spells.Ready && con.spells.BlinkReady && con.unit != con.team.TeamTank {
		// Trigger cooldown.
		con.spells.Ready = false
		// Check reaction time based on difficulty setting.
		ns.NewTimer(ns.Frames(con.reactionTime), func() {
			// Check for War Cry before chant.
			if con.spells.isAlive && !con.unit.HasEnchant(enchant.ANTI_MAGIC) {
				castPhonemes(con.unit, []audio.Name{PhRight, PhLeft, PhUp}, func() {
					// Check for War Cry before spell release.
					if con.spells.isAlive && !con.unit.HasEnchant(enchant.ANTI_MAGIC) {
						con.spells.BlinkReady = false
						con.mana = con.mana - 10
						ns.NewTrap(con.unit, spell.BLINK)
						// Global cooldown.
						ns.NewTimer(ns.Frames(3), func() {
							con.spells.Ready = true
						})
						// Blink cooldown.
						ns.NewTimer(ns.Seconds(1), func() {
							con.spells.BlinkReady = true
						})
					}
				})
			} else {
				ns.NewTimer(ns.Frames(con.reactionTime), func() {
					con.spells.Ready = true
				})
			}
		})
	}
}

func (con *Conjurer) castBurn() {
	// Check if cooldowns are ready.
	if con.mana >= 10 && con.spells.isAlive && !con.unit.HasEnchant(enchant.ANTI_MAGIC) && !con.target.HasEnchant(enchant.INVULNERABLE) && con.spells.BurnReady && con.spells.Ready && con.target.HasEnchant(enchant.REFLECTIVE_SHIELD) && !con.target.HasEnchant(enchant.INVULNERABLE) {
		// Select target.
		con.cursor = con.target.Pos()
		// Trigger cooldown.
		con.spells.Ready = false
		// Check reaction time based on difficulty setting.
		ns.NewTimer(ns.Frames(con.reactionTime), func() {
			// Check for War Cry before chant.
			if con.spells.isAlive && !con.unit.HasEnchant(enchant.ANTI_MAGIC) {
				castPhonemes(con.unit, []audio.Name{PhDown, PhDown, PhUp, PhUp}, func() {
					// Check for War Cry before spell release.
					if con.spells.isAlive && !con.unit.HasEnchant(enchant.ANTI_MAGIC) && con.unit.CanSee(con.target) {
						// Aim.
						con.unit.LookAtObject(con.target)
						con.unit.Pause(ns.Frames(con.reactionTime))
						con.spells.BurnReady = false
						ns.CastSpell(spell.BURN, con.unit, con.cursor)
						con.mana = con.mana - 10
						// Global cooldown.
						ns.NewTimer(ns.Frames(3), func() {
							con.spells.Ready = true
						})
						// Burn cooldown.
						ns.NewTimer(ns.Frames(3), func() {
							con.spells.BurnReady = true
						})
					}
				})
			} else {
				ns.NewTimer(ns.Frames(con.reactionTime), func() {
					con.spells.Ready = true
				})
			}
		})
	}
}

func (con *Conjurer) castStun() {
	// Check if cooldowns are ready.
	if con.spells.isAlive && !con.unit.HasEnchant(enchant.ANTI_MAGIC) && con.unit.CanSee(con.target) && con.spells.StunReady && con.spells.Ready && !con.target.HasEnchant(enchant.HELD) && !con.target.HasEnchant(enchant.SLOWED) && con.target.MaxHealth() != 150 {
		// Select target.
		// Trigger cooldown.
		con.spells.Ready = false
		// Check reaction time based on difficulty setting.
		ns.NewTimer(ns.Frames(con.reactionTime), func() {
			// Check for War Cry before chant.
			if con.spells.isAlive && !con.unit.HasEnchant(enchant.ANTI_MAGIC) {
				castPhonemes(con.unit, []audio.Name{PhUpLeft, PhDown}, func() {
					// Check for War Cry before spell release.
					if con.spells.isAlive && !con.unit.HasEnchant(enchant.ANTI_MAGIC) {
						// Aim.
						con.unit.LookAtObject(con.target)
						con.unit.Pause(ns.Frames(con.reactionTime))
						con.spells.StunReady = false
						ns.CastSpell(spell.STUN, con.unit, con.target)
						// Global cooldown.
						con.spells.Ready = true
						ns.NewTimer(ns.Seconds(5), func() {
							// Stun cooldown.
							con.spells.StunReady = true
						})
					}
				})
			}
		})
	}
}

func (con *Conjurer) castToxicCloud() {
	// Check if cooldowns are ready.
	if con.mana >= 60 && con.spells.isAlive && !con.unit.HasEnchant(enchant.ANTI_MAGIC) && con.unit.CanSee(con.target) && con.spells.ToxicCloudReady && con.spells.Ready {
		// Select target.
		con.cursor = con.target.Pos()
		// Trigger cooldown.
		con.spells.Ready = false
		// Check reaction time based on difficulty setting.
		ns.NewTimer(ns.Frames(con.reactionTime), func() {
			// Check for War Cry before chant.
			if con.spells.isAlive && !con.unit.HasEnchant(enchant.ANTI_MAGIC) {
				castPhonemes(con.unit, []audio.Name{PhUpRight, PhDownLeft, PhUpLeft}, func() {
					// Check for War Cry before spell release.
					if con.spells.isAlive && !con.unit.HasEnchant(enchant.ANTI_MAGIC) {
						// Aim.
						con.unit.LookAtObject(con.target)
						con.unit.Pause(ns.Frames(con.reactionTime))
						con.spells.ToxicCloudReady = false
						ns.CastSpell(spell.TOXIC_CLOUD, con.unit, con.cursor)
						con.mana = con.mana - 60
						// Global cooldown.
						con.spells.Ready = true
						// Toxic Cloud cooldown.
						ns.NewTimer(ns.Seconds(10), func() {
							con.spells.ToxicCloudReady = true
						})
					}
				})
			}
		})
	}
}

func (con *Conjurer) castSlow() {
	// Check if cooldowns are ready.
	if con.mana >= 10 && con.spells.isAlive && !con.unit.HasEnchant(enchant.ANTI_MAGIC) && con.unit.CanSee(con.target) && con.spells.SlowReady && con.spells.Ready && !con.target.HasEnchant(enchant.SLOWED) && !con.target.HasEnchant(enchant.REFLECTIVE_SHIELD) && !con.target.HasEnchant(enchant.HELD) {
		// Select target.
		// Trigger cooldown.
		con.spells.Ready = false
		// Check reaction time based on difficulty setting.
		ns.NewTimer(ns.Frames(con.reactionTime), func() {
			// Check for War Cry before chant.
			if con.spells.isAlive && !con.unit.HasEnchant(enchant.ANTI_MAGIC) {
				castPhonemes(con.unit, []audio.Name{PhDown, PhDown, PhDown}, func() {
					// Check for War Cry before spell release.
					if con.spells.isAlive && !con.unit.HasEnchant(enchant.ANTI_MAGIC) {
						// Aim.
						con.unit.LookAtObject(con.target)
						con.unit.Pause(ns.Frames(con.reactionTime))
						con.spells.SlowReady = false
						con.mana = con.mana - 10
						ns.CastSpell(spell.SLOW, con.unit, con.target)
						// Global cooldown.
						ns.NewTimer(ns.Frames(3), func() {
							con.spells.Ready = true
						})
						// Slow cooldown.
						ns.NewTimer(ns.Seconds(3), func() {
							con.spells.SlowReady = true
						})
					}
				})
			} else {
				ns.NewTimer(ns.Frames(con.reactionTime), func() {
					con.spells.Ready = true
				})
			}
		})
	}
}

func (con *Conjurer) castCounterspell() {
	// Check if cooldowns are ready.
	if con.mana >= 20 && con.spells.isAlive && !con.unit.HasEnchant(enchant.ANTI_MAGIC) && !con.unit.HasEnchant(enchant.INVISIBLE) && con.target.HasEnchant(enchant.SHOCK) && con.spells.Ready && con.spells.CounterspellReady && con.unit.CanSee(con.target) {
		// Trigger cooldown.
		con.spells.Ready = false
		// Check reaction time based on difficulty setting.
		ns.NewTimer(ns.Frames(con.reactionTime), func() {
			// Check for War Cry before chant.
			if con.spells.isAlive && !con.unit.HasEnchant(enchant.ANTI_MAGIC) {
				castPhonemes(con.unit, []audio.Name{PhDown, PhDownRight}, func() {
					// Check for War Cry before spell release.
					if con.spells.isAlive && !con.unit.HasEnchant(enchant.ANTI_MAGIC) && con.unit.CanSee(con.target) {
						con.spells.CounterspellReady = false
						con.mana = con.mana - 20
						ns.CastSpell(spell.COUNTERSPELL, con.unit, con.unit.Pos())
						// Global cooldown.
						ns.NewTimer(ns.Frames(3), func() {
							con.spells.Ready = true
						})
						// Haste cooldown.
						ns.NewTimer(ns.Seconds(20), func() {
							con.spells.CounterspellReady = true
						})
					}
				})
			} else {
				ns.NewTimer(ns.Frames(con.reactionTime), func() {
					con.spells.Ready = true
				})
			}
		})
	}
}

func (con *Conjurer) castMeteor() {
	// Check if cooldowns are ready.
	if con.mana >= 30 && con.spells.isAlive && !con.unit.HasEnchant(enchant.ANTI_MAGIC) && con.unit.CanSee(con.target) && con.spells.MeteorReady && con.spells.Ready {
		// Select target.
		con.cursor = con.target.Pos()
		// Trigger cooldown.
		con.spells.Ready = false
		// Check reaction time based on difficulty setting.
		ns.NewTimer(ns.Frames(con.reactionTime), func() {
			// Check for War Cry before chant.
			if con.spells.isAlive && !con.unit.HasEnchant(enchant.ANTI_MAGIC) {
				castPhonemes(con.unit, []audio.Name{PhDownLeft, PhDownLeft}, func() {
					// Check for War Cry before spell release.
					if con.spells.isAlive && !con.unit.HasEnchant(enchant.ANTI_MAGIC) {
						// Aim.
						con.unit.LookAtObject(con.target)
						con.unit.Pause(ns.Frames(con.reactionTime))
						con.spells.MeteorReady = false
						ns.CastSpell(spell.METEOR, con.unit, con.cursor)
						con.mana = con.mana - 30
						// Global cooldown.
						con.spells.Ready = true
						ns.NewTimer(ns.Seconds(10), func() {
							// Meteor cooldown.
							con.spells.MeteorReady = true
						})
					}
				})
			}
		})
	}
}

func (con *Conjurer) castInfravision() {
	// Check if cooldowns are ready.
	if con.mana >= 30 && con.spells.isAlive && !con.unit.HasEnchant(enchant.ANTI_MAGIC) && con.spells.Ready && con.spells.InfravisionReady {
		// Trigger cooldown.
		con.spells.Ready = false
		// Check reaction time based on difficulty setting.
		ns.NewTimer(ns.Frames(con.reactionTime), func() {
			// Check for War Cry before chant.
			if con.spells.isAlive && !con.unit.HasEnchant(enchant.ANTI_MAGIC) {
				castPhonemes(con.unit, []audio.Name{PhRight, PhLeft, PhRight, PhLeft}, func() {
					// Check for War Cry before spell release.
					if con.spells.isAlive && !con.unit.HasEnchant(enchant.ANTI_MAGIC) {
						con.spells.InfravisionReady = false
						ns.CastSpell(spell.INFRAVISION, con.unit, con.unit)
						con.mana = con.mana - 30
						// Global cooldown.
						con.spells.Ready = true
						// Invravision cooldown.
						ns.NewTimer(ns.Seconds(30), func() {
							con.spells.InfravisionReady = true
						})
					}
				})
			}
		})
	}
}

//func (con *Conjurer) summonGhost() {
//	// Check if cooldowns are ready.
//	if con.spells.isAlive && !con.unit.HasEnchant(enchant.ANTI_MAGIC) && con.spells.Ready && con.spells.SummonGhostReady {
//		// Trigger cooldown.
//		con.spells.Ready = false
//		// Check reaction time based on difficulty setting.
//		ns.NewTimer(ns.Frames(con.reactionTime), func() {
//			// Check for War Cry before chant.
//			if con.spells.isAlive && !con.unit.HasEnchant(enchant.ANTI_MAGIC) {
//				castPhonemes(con.unit, []audio.Name{PhUpLeft, PhDownRight, PhUpRight, PhDownLeft, PhDown}, func() {
//					// Check for War Cry before spell release.
//					if con.spells.isAlive && !con.unit.HasEnchant(enchant.ANTI_MAGIC) {
//						con.spells.SummonGhostReady = false
//						ns.CastSpell(spell.SUMMON_GHOST, con.unit, con.unit)
//						// Global cooldown.
//						con.spells.Ready = true
//						// Summon Ghost cooldown.
//						ns.NewTimer(ns.Seconds(5), func() {
//							con.spells.SummonGhostReady = true
//						})
//					}
//				})
//			}
//		})
//	}
//}

func (con *Conjurer) summonBomber1() {
	// Check if cooldowns are ready.
	if con.spells.isAlive && !con.unit.HasEnchant(enchant.ANTI_MAGIC) && con.spells.Ready && con.spells.SummonBomber1Ready {
		// Trigger cooldown.
		con.spells.Ready = false
		// Check reaction time based on difficulty setting.
		ns.NewTimer(ns.Frames(con.reactionTime), func() {
			// Check for War Cry before chant.
			if con.spells.isAlive && !con.unit.HasEnchant(enchant.ANTI_MAGIC) && GameModeIsCTF {
				// Slow chant.
				castPhonemes(con.unit, []audio.Name{PhDown, PhDown, PhDown}, func() {
					// Pause for concentration.
					ns.NewTimer(ns.Frames(3), func() {
						// Check for War Cry before chant.
						if con.spells.isAlive && !con.unit.HasEnchant(enchant.ANTI_MAGIC) {
							// Poison chant.
							castPhonemes(con.unit, []audio.Name{PhUpRight, PhDownLeft}, func() {
								// Pause for concentration.
								ns.NewTimer(ns.Frames(3), func() {
									// Check for War Cry before chant.
									if con.spells.isAlive && !con.unit.HasEnchant(enchant.ANTI_MAGIC) {
										// Fist Of Vengeance chant.
										castPhonemes(con.unit, []audio.Name{PhUpRight, PhUp, PhDown}, func() {
											// Pause for concentration.
											ns.NewTimer(ns.Frames(3), func() {
												// Check for War Cry before chant.
												if con.spells.isAlive && !con.unit.HasEnchant(enchant.ANTI_MAGIC) {
													// Glyph chant.
													castPhonemes(con.unit, []audio.Name{PhUp, PhRight, PhLeft, PhDown}, func() {
														// Check for War Cry before spell release.
														if con.spells.isAlive && !con.unit.HasEnchant(enchant.ANTI_MAGIC) {
															con.spells.SummonBomber1Ready = false
															con.bomber1 = ns.CreateObject("Bomber", con.unit)
															ns.AudioEvent("BomberSummon", con.bomber1)
															con.bomber1.SetOwner(con.unit)
															con.bomber1.SetTeam(con.team.Team())
															con.bomber1.OnEvent(ns.ObjectEvent(ns.EventDeath), func() {
																// Summon Bomber cooldown.
																ns.NewTimer(ns.Seconds(10), func() {
																	con.spells.SummonBomber1Ready = true
																})
															})
															con.bomber1.Follow(con.unit)
															con.bomber1.TrapSpells(spell.POISON, spell.FIST, spell.SLOW)
															con.bomber1.OnEvent(ns.ObjectEvent(ns.EventEnemySighted), func() {
																con.bomber1.Attack(con.target)
															})
															con.bomber1.OnEvent(ns.ObjectEvent(ns.EventEnemyHeard), func() {
																con.bomber1.Attack(con.target)
															})
															con.bomber1.OnEvent(ns.ObjectEvent(ns.EventLostEnemy), func() {
																con.bomber1.Follow(con.unit)
															})
															// Global cooldown.
															con.spells.Ready = true
														}
													})
												}
											})
										})
									}
								})
							})
						}
					})
				})
			} else if con.spells.isAlive && !con.unit.HasEnchant(enchant.ANTI_MAGIC) && !GameModeIsCTF {
				// Stun chant.
				castPhonemes(con.unit, []audio.Name{PhUpLeft, PhDown}, func() {
					// Pause for concentration.
					ns.NewTimer(ns.Frames(3), func() {
						// Check for War Cry before chant.
						if con.spells.isAlive && !con.unit.HasEnchant(enchant.ANTI_MAGIC) {
							// Poison chant.
							castPhonemes(con.unit, []audio.Name{PhUpRight, PhDownLeft}, func() {
								// Pause for concentration.
								ns.NewTimer(ns.Frames(3), func() {
									// Check for War Cry before chant.
									if con.spells.isAlive && !con.unit.HasEnchant(enchant.ANTI_MAGIC) {
										// Fist Of Vengeance chant.
										castPhonemes(con.unit, []audio.Name{PhUpRight, PhUp, PhDown}, func() {
											// Pause for concentration.
											ns.NewTimer(ns.Frames(3), func() {
												// Check for War Cry before chant.
												if con.spells.isAlive && !con.unit.HasEnchant(enchant.ANTI_MAGIC) {
													// Glyph chant.
													castPhonemes(con.unit, []audio.Name{PhUp, PhRight, PhLeft, PhDown}, func() {
														// Check for War Cry before spell release.
														if con.spells.isAlive && !con.unit.HasEnchant(enchant.ANTI_MAGIC) {
															con.spells.SummonBomber1Ready = false
															con.bomber1 = ns.CreateObject("Bomber", con.unit)
															ns.AudioEvent("BomberSummon", con.bomber1)
															con.bomber1.SetOwner(con.unit)
															con.bomber1.SetTeam(con.team.Team())
															con.bomber1.OnEvent(ns.ObjectEvent(ns.EventDeath), func() {
																// Summon Bomber cooldown.
																ns.NewTimer(ns.Seconds(10), func() {
																	con.spells.SummonBomber1Ready = true
																})
															})
															con.bomber1.Follow(con.unit)
															con.bomber1.TrapSpells(spell.POISON, spell.FIST, spell.STUN)
															con.bomber1.OnEvent(ns.ObjectEvent(ns.EventEnemySighted), func() {
																con.bomber1.Attack(con.target)
															})
															con.bomber1.OnEvent(ns.ObjectEvent(ns.EventEnemyHeard), func() {
																con.bomber1.Attack(con.target)
															})
															con.bomber1.OnEvent(ns.ObjectEvent(ns.EventLostEnemy), func() {
																con.bomber1.Follow(con.unit)
															})
															// Global cooldown.
															con.spells.Ready = true
														}
													})
												}
											})
										})
									}
								})
							})
						}
					})
				})
			}
		})
	}
}

func (con *Conjurer) castCounterspellAtForceOfNature() {
	// Check if cooldowns are ready.
	if con.mana >= 20 && con.spells.isAlive && !con.unit.HasEnchant(enchant.ANTI_MAGIC) && con.spells.Ready && con.spells.CounterspellReady {
		// Trigger cooldown.
		con.spells.Ready = false
		// Check reaction time based on difficulty setting.
		ns.NewTimer(ns.Frames(con.reactionTime), func() {
			// Check for War Cry before chant.
			if con.spells.isAlive && !con.unit.HasEnchant(enchant.ANTI_MAGIC) {
				castPhonemes(con.unit, []audio.Name{PhDown, PhDownRight}, func() {
					// Check for War Cry before spell release.
					if con.spells.isAlive && !con.unit.HasEnchant(enchant.ANTI_MAGIC) {
						con.spells.CounterspellReady = false
						con.mana = con.mana - 20
						ns.CastSpell(spell.COUNTERSPELL, con.unit, con.unit.Pos())
						// Global cooldown.
						ns.NewTimer(ns.Frames(3), func() {
							con.spells.Ready = true
						})
						// Haste cooldown.
						ns.NewTimer(ns.Seconds(20), func() {
							con.spells.CounterspellReady = true
						})
					}
				})
			} else {
				ns.NewTimer(ns.Frames(con.reactionTime), func() {
					con.spells.Ready = true
				})
			}
		})
	}
}

func (con *Conjurer) summonBomber2() {
	// Check if cooldowns are ready.
	if con.spells.isAlive && !con.unit.HasEnchant(enchant.ANTI_MAGIC) && con.spells.Ready && con.spells.SummonBomber2Ready {
		// Trigger cooldown.
		con.spells.Ready = false
		// Check reaction time based on difficulty setting.
		ns.NewTimer(ns.Frames(con.reactionTime), func() {
			// Check for War Cry before chant.
			if con.spells.isAlive && !con.unit.HasEnchant(enchant.ANTI_MAGIC) && GameModeIsCTF {
				// Slow chant.
				castPhonemes(con.unit, []audio.Name{PhDown, PhDown, PhDown}, func() {
					// Pause for concentration.
					ns.NewTimer(ns.Frames(3), func() {
						// Check for War Cry before chant.
						if con.spells.isAlive && !con.unit.HasEnchant(enchant.ANTI_MAGIC) {
							// Poison chant.
							castPhonemes(con.unit, []audio.Name{PhUpRight, PhDownLeft}, func() {
								// Pause for concentration.
								ns.NewTimer(ns.Frames(3), func() {
									// Check for War Cry before chant.
									if con.spells.isAlive && !con.unit.HasEnchant(enchant.ANTI_MAGIC) {
										// Fist Of Vengeance chant.
										castPhonemes(con.unit, []audio.Name{PhUpRight, PhUp, PhDown}, func() {
											// Pause for concentration.
											ns.NewTimer(ns.Frames(3), func() {
												// Check for War Cry before chant.
												if con.spells.isAlive && !con.unit.HasEnchant(enchant.ANTI_MAGIC) {
													// Glyph chant.
													castPhonemes(con.unit, []audio.Name{PhUp, PhRight, PhLeft, PhDown}, func() {
														// Check for War Cry before spell release.
														if con.spells.isAlive && !con.unit.HasEnchant(enchant.ANTI_MAGIC) {
															con.spells.SummonBomber2Ready = false
															con.bomber2 = ns.CreateObject("Bomber", con.unit)
															ns.AudioEvent("BomberSummon", con.bomber2)
															con.bomber2.SetOwner(con.unit)
															con.bomber2.SetTeam(con.team.Team())
															con.bomber2.OnEvent(ns.ObjectEvent(ns.EventDeath), func() {
																// Summon Bomber cooldown.
																ns.NewTimer(ns.Seconds(10), func() {
																	con.spells.SummonBomber2Ready = true
																})
															})
															con.bomber2.Follow(con.unit)
															con.bomber2.TrapSpells(spell.POISON, spell.FIST, spell.SLOW)
															con.bomber2.OnEvent(ns.ObjectEvent(ns.EventEnemySighted), func() {
																con.bomber2.Attack(con.target)
															})
															con.bomber2.OnEvent(ns.ObjectEvent(ns.EventEnemyHeard), func() {
																con.bomber2.Attack(con.target)
															})
															con.bomber2.OnEvent(ns.ObjectEvent(ns.EventLostEnemy), func() {
																con.bomber2.Follow(con.unit)
															})
															// Global cooldown.
															con.spells.Ready = true
														}
													})
												}
											})
										})
									}
								})
							})
						}
					})
				})
			} else if con.spells.isAlive && !con.unit.HasEnchant(enchant.ANTI_MAGIC) && !GameModeIsCTF {
				// Stun chant.
				castPhonemes(con.unit, []audio.Name{PhUpLeft, PhDown}, func() {
					// Pause for concentration.
					ns.NewTimer(ns.Frames(3), func() {
						// Check for War Cry before chant.
						if con.spells.isAlive && !con.unit.HasEnchant(enchant.ANTI_MAGIC) {
							// Poison chant.
							castPhonemes(con.unit, []audio.Name{PhUpRight, PhDownLeft}, func() {
								// Pause for concentration.
								ns.NewTimer(ns.Frames(3), func() {
									// Check for War Cry before chant.
									if con.spells.isAlive && !con.unit.HasEnchant(enchant.ANTI_MAGIC) {
										// Fist Of Vengeance chant.
										castPhonemes(con.unit, []audio.Name{PhUpRight, PhUp, PhDown}, func() {
											// Pause for concentration.
											ns.NewTimer(ns.Frames(3), func() {
												// Check for War Cry before chant.
												if con.spells.isAlive && !con.unit.HasEnchant(enchant.ANTI_MAGIC) {
													// Glyph chant.
													castPhonemes(con.unit, []audio.Name{PhUp, PhRight, PhLeft, PhDown}, func() {
														// Check for War Cry before spell release.
														if con.spells.isAlive && !con.unit.HasEnchant(enchant.ANTI_MAGIC) {
															con.spells.SummonBomber2Ready = false
															con.bomber2 = ns.CreateObject("Bomber", con.unit)
															ns.AudioEvent("BomberSummon", con.bomber2)
															con.bomber2.SetOwner(con.unit)
															con.bomber2.SetTeam(con.team.Team())
															con.bomber2.OnEvent(ns.ObjectEvent(ns.EventDeath), func() {
																// Summon Bomber cooldown.
																ns.NewTimer(ns.Seconds(10), func() {
																	con.spells.SummonBomber2Ready = true
																})
															})
															con.bomber2.Follow(con.unit)
															con.bomber2.TrapSpells(spell.POISON, spell.FIST, spell.STUN)
															con.bomber2.OnEvent(ns.ObjectEvent(ns.EventEnemySighted), func() {
																con.bomber2.Attack(con.target)
															})
															con.bomber2.OnEvent(ns.ObjectEvent(ns.EventEnemyHeard), func() {
																con.bomber2.Attack(con.target)
															})
															con.bomber2.OnEvent(ns.ObjectEvent(ns.EventLostEnemy), func() {
																con.bomber2.Follow(con.unit)
															})
															// Global cooldown.
															con.spells.Ready = true
														}
													})
												}
											})
										})
									}
								})
							})
						}
					})
				})
			}
		})
	}
}

// ------------------------------------------------------------------------------------------------------------------------------------ //
// ---------------------------------------------------------------- SPELL BOOK -------------------------------------------------------- //
// ------------------------------------------------------------------------------------------------------------------------------------ //

func (con *Conjurer) onConCommand(t ns.Team, p ns.Player, obj ns.Obj, msg string) string {
	if p != nil {
		switch msg {
		// Spawn commands red bots.
		case "vamp", "Vamp", "Vampirism", "vampirism":
			if con.unit.CanSee(p.Unit()) && con.unit.HasTeam(p.Unit().Team()) {
				if con.mana >= 20 && con.spells.isAlive && !con.unit.HasEnchant(enchant.ANTI_MAGIC) && con.spells.Ready {
					// Trigger cooldown.
					con.spells.Ready = false
					// Check reaction time based on difficulty setting.
					ns.NewTimer(ns.Frames(con.reactionTime), func() {
						// Check for War Cry before chant.
						if con.spells.isAlive && !con.unit.HasEnchant(enchant.ANTI_MAGIC) {
							castPhonemes(con.unit, []audio.Name{PhUp, PhDown, PhLeft, PhRight}, func() {
								// Check for War Cry before spell release.
								if con.spells.isAlive && !con.unit.HasEnchant(enchant.ANTI_MAGIC) {
									con.spells.VampirismReady = false
									con.mana = con.mana - 20
									ns.CastSpell(spell.VAMPIRISM, con.unit, p.Unit())
									// Global cooldown.
									con.spells.Ready = true
									// Vampirism cooldown.
									ns.NewTimer(ns.Seconds(30), func() {
										con.spells.VampirismReady = true
									})
								}
							})
						}
					})
				}
			}
			if con.mana < 20 && con.spells.isAlive && !con.unit.HasEnchant(enchant.ANTI_MAGIC) && con.spells.Ready {
				// Trigger cooldown.
				con.spells.Ready = false
				// Check reaction time based on difficulty setting.
				ns.NewTimer(ns.Frames(con.reactionTime), func() {
					// Check for War Cry before chant.
					if con.spells.isAlive && !con.unit.HasEnchant(enchant.ANTI_MAGIC) {
						castPhonemes(con.unit, []audio.Name{PhUp, PhDown, PhLeft, PhRight}, func() {
							// Check for War Cry before spell release.
							if con.spells.isAlive && !con.unit.HasEnchant(enchant.ANTI_MAGIC) {
								con.spells.VampirismReady = false
								ns.AudioEvent(audio.ManaEmpty, con.unit)
								// Global cooldown.
								con.spells.Ready = true
								// Vampirism cooldown.
								ns.NewTimer(ns.Seconds(30), func() {
									con.spells.VampirismReady = true
								})
							}
						})
					}
				})
			}
		}
	}
	return msg
}
