package main

import "strings"

type player struct {
	location string
	haveItem map[string]bool
	action   map[string]func([]string) string
}
type room struct {
	description     string
	events          map[string]bool
	havePathTo      []string
	placesWithItems [][]string
	additionTo      map[string]func([]string) string
}

func lookAround(args []string) string {
	const actionType = "осмотреться"
	room := rooms[vasya.location]
	var roomItemsInfo, roomGateInfo string

	var roomItems []string
	for _, area := range room.placesWithItems {
		place, items := area[0], area[1:]

		var filteredItems []string
		for _, item := range items {
			if !vasya.haveItem[item] {
				filteredItems = append(filteredItems, item)
			}
		}
		if len(filteredItems) > 0 {
			placeItemsInfo := place + ": " + strings.Join(filteredItems, ", ")
			roomItems = append(roomItems, placeItemsInfo)
		}
	}

	if len(roomItems) > 0 {
		roomItemsInfo = strings.Join(roomItems, ", ")
	} else {
		roomItemsInfo = "пустая комната"
	}
	roomGateInfo = "можно пройти - " + strings.Join(room.havePathTo, ", ")

	if add, haveAdd := room.additionTo[actionType]; haveAdd {
		return add([]string{roomItemsInfo, roomGateInfo})
	}

	return roomItemsInfo + ". " + roomGateInfo
}
func goTo(args []string) string {
	const actionType = "идти"
	room := rooms[vasya.location]
	destination := args[0]

	var havePathToDest bool
	for _, path := range room.havePathTo {
		if path == destination {
			havePathToDest = true
			break
		}
	}

	if havePathToDest {
		if add, haveAdd := room.additionTo[actionType]; haveAdd {
			res := add([]string{destination})
			if res != "" {
				return res
			}
		}

		vasya.location = destination
		room := rooms[vasya.location]

		availibleRooms := strings.Join(room.havePathTo, ", ")
		return room.description + ". " + "можно пройти - " + availibleRooms
	}

	return "нет пути в " + destination
}
func takeIt(args []string) string {
	room := rooms[vasya.location]
	requiredItem := args[0]

	if !vasya.haveItem["рюкзак"] {
		return "некуда класть"
	}

	var itemInRoom, playerHaveItem bool
Loop:
	for _, place := range room.placesWithItems {
		for _, item := range place {
			if item == requiredItem {
				itemInRoom = true
				break Loop
			}
		}
	}
	playerHaveItem = vasya.haveItem[requiredItem]

	if itemInRoom && !playerHaveItem {
		vasya.haveItem[requiredItem] = true

		return "предмет добавлен в инвентарь: " + requiredItem
	}

	return "нет такого"
}
func putOn(args []string) string {
	const actionType = "надеть"
	room := rooms[vasya.location]
	item := args[0]

	if add, haveAdd := room.additionTo[actionType]; haveAdd {
		ans := add([]string{item})
		if ans != "" {
			return ans
		}
	}
	return "не к чему применить"
}
func applyTo(args []string) string {
	const actionType = "применить"
	room := rooms[vasya.location]
	tool, application := args[0], args[1]

	if !vasya.haveItem[tool] {
		return "нет предмета в инвентаре - " + tool
	}

	if add, haveAdd := room.additionTo[actionType]; haveAdd {
		ans := add([]string{tool, application})
		if ans != "" {
			return ans
		}
	}

	return "не к чему применить"
}

var vasya player
var rooms map[string]room

func initGame() {
	vasya = player{
		location: "кухня",
		haveItem: make(map[string]bool),
		action: map[string]func([]string) string{
			"осмотреться": lookAround,
			"идти":        goTo,
			"взять":       takeIt,
			"надеть":      putOn,
			"применить":   applyTo,
		},
	}

	var kitchen = room{
		events:          map[string]bool{},
		description:     "кухня, ничего интересного",
		havePathTo:      []string{"коридор"},
		placesWithItems: [][]string{{"на столе", "чай"}},
		additionTo: map[string]func([]string) string{
			"осмотреться": func(args []string) string {
				roomItemsInfo, roomGateInfo := args[0], args[1]

				var planInfo string
				if vasya.haveItem["рюкзак"] {
					planInfo = ", надо идти в универ. "
				} else {
					planInfo = ", надо собрать рюкзак и идти в универ. "
				}
				return "ты находишься на кухне, " +
					roomItemsInfo + planInfo + roomGateInfo
			},
		},
	}
	var hallway = room{
		events:          map[string]bool{},
		description:     "ничего интересного",
		havePathTo:      []string{"кухня", "комната", "улица"},
		placesWithItems: [][]string{},
		additionTo: map[string]func([]string) string{
			"идти": func(args []string) string {
				destination := args[0]
				room := rooms["коридор"]

				if destination == "улица" && !room.events["дверь на улицу открыта"] {
					return "дверь закрыта"
				}
				return ""
			},
			"применить": func(args []string) string {
				room := rooms["коридор"]
				tool, application := args[0], args[1]

				if tool == "ключи" && application == "дверь" {
					room.events["дверь на улицу открыта"] = true
					return "дверь открыта"
				}
				return ""
			},
		},
	}
	var bedroom = room{
		events:      map[string]bool{},
		description: "ты в своей комнате",
		havePathTo:  []string{"коридор"},
		placesWithItems: [][]string{
			{"на столе", "ключи", "конспекты"},
			{"на стуле", "рюкзак"},
		},
		additionTo: map[string]func([]string) string{
			"надеть": func(args []string) string {
				item := args[0]

				if item == "рюкзак" {
					vasya.haveItem["рюкзак"] = true
					return "вы надели: рюкзак"
				}
				return ""
			},
		},
	}
	var street = room{
		events:          map[string]bool{},
		description:     "на улице весна",
		havePathTo:      []string{"домой"},
		placesWithItems: [][]string{},
	}

	rooms = map[string]room{
		"кухня":   kitchen,
		"коридор": hallway,
		"комната": bedroom,
		"улица":   street,
	}
}

func handleCommand(command string) string {
	input := strings.Split(command, " ")
	act, args := input[0], input[1:]

	_, actionExist := vasya.action[act]
	if actionExist {
		return vasya.action[act](args)
	}
	return "неизвестная команда"
}
