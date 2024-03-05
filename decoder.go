package sxtp

import (
	"bufio"
	"fmt"
	"image"
	"io"
	"reflect"
	"strconv"
	"strings"
)

func typeOf[T any]() reflect.Type {
	return reflect.TypeFor[T]()
}

var parsers = map[reflect.Type]func(string) (reflect.Value, error){
	typeOf[string](): func(s string) (reflect.Value, error) {
		return reflect.ValueOf(s), nil
	},
	typeOf[bool](): func(s string) (reflect.Value, error) {
		v, err := strconv.ParseBool(s)
		return reflect.ValueOf(v), err
	},
	typeOf[int](): func(s string) (reflect.Value, error) {
		v, err := strconv.Atoi(s)
		return reflect.ValueOf(v), err
	},
	typeOf[Filter](): func(s string) (reflect.Value, error) {
		split := strings.SplitN(s, ",", 2)
		return reflect.ValueOf(Filter{X: strings.TrimSpace(split[0]), Y: strings.TrimSpace(split[1])}), nil
	},
	typeOf[Angle](): func(s string) (reflect.Value, error) {
		value, err := strconv.ParseBool(s)
		if err == nil {
			if value {
				return reflect.ValueOf(Angle(90)), nil
			} else {
				return reflect.ValueOf(Angle(0)), nil
			}
		}
		floatValue, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return reflect.ValueOf(nil), err
		}
		return reflect.ValueOf(Angle(floatValue)), nil
	},
	typeOf[image.Point](): func(s string) (reflect.Value, error) {
		components, err := parseInts(s, 2)
		if err != nil {
			return reflect.ValueOf(nil), err
		}
		return reflect.ValueOf(image.Pt(components[0], components[1])), nil
	},
	typeOf[Bounds](): func(s string) (reflect.Value, error) {
		components, err := parseInts(s, 4)
		if err != nil {
			return reflect.ValueOf(nil), err
		}
		return reflect.ValueOf(Bounds{
			Position: image.Pt(components[0], components[1]),
			Size:     image.Pt(components[2], components[3]),
		}), nil
	},
	typeOf[Offsets](): func(s string) (reflect.Value, error) {
		components, err := parseInts(s, 4)
		if err != nil {
			return reflect.ValueOf(nil), err
		}
		return reflect.ValueOf(Offsets{
			Offset:       image.Pt(components[0], components[1]),
			OriginalSize: image.Pt(components[2], components[3]),
		}), nil
	},
}

// DecodeAtlas Parses a spine atlas file line by line from an io.Reader and returns a struct describing the layout
func DecodeAtlas(reader io.Reader) ([]Atlas, error) {
	scanner := bufio.NewScanner(reader)
	scanner.Split(bufio.ScanLines)

	processedHeader := false
	openSpriteBlock := false

	pages := make([]Atlas, 0)
	var currentAtlas *Atlas

	atlasType := reflect.TypeOf(currentAtlas).Elem()
	atlasValue := reflect.ValueOf(currentAtlas).Elem()

	sprites := make([]Sprite, 0)
	currentSprite := new(Sprite)

	spriteType := reflect.TypeOf(*currentSprite)
	currentSpriteValue := reflect.ValueOf(currentSprite).Elem()

	for scanner.Scan() {
		text := scanner.Text()

		// Open Page
		if text == "" || currentAtlas == nil {
			if currentAtlas != nil {
				pages = append(pages, *currentAtlas)
			}

			currentAtlas = new(Atlas)
			if text != "" {
				currentAtlas.Name = text
			}
			atlasValue = reflect.ValueOf(currentAtlas).Elem()

			processedHeader = false
			openSpriteBlock = false
			continue
		}

		// open currentAtlas block
		if currentAtlas.Name == "" && !strings.Contains(text, ":") {
			currentAtlas.Name = text
			continue
		}

		// open sprite block
		if !strings.Contains(text, ":") {
			if openSpriteBlock {
				setSpriteDefaults(currentSprite)
				sprites = append(sprites, *currentSprite)
				currentSprite = new(Sprite)
				currentSpriteValue = reflect.ValueOf(currentSprite).Elem()
			}

			processedHeader = true
			openSpriteBlock = true
			currentSprite.Name = text
			continue
		}

		if processedHeader {
			if err := parseLine(text, spriteType, currentSpriteValue); err != nil {
				return nil, err
			}
		} else if err := parseLine(text, atlasType, atlasValue); err != nil {
			return nil, err
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if currentAtlas != nil {
		setSpriteDefaults(currentSprite)
		sprites = append(sprites, *currentSprite)
		currentAtlas.Sprites = sprites
		pages = append(pages, *currentAtlas)
	}

	return pages, nil
}

func parseLine(text string, t reflect.Type, v reflect.Value) error {
	key, value := kvp(text)

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("atlas")

		if key != tag {
			continue
		}

		parsed, err := parseValue(value, field.Type)
		if err != nil {
			return err
		}

		v.Field(i).Set(parsed)
	}
	return nil
}

func parseValue(value string, fieldType reflect.Type) (reflect.Value, error) {
	parser, ok := parsers[fieldType]
	if ok {
		return parser(value)
	}
	return reflect.ValueOf(nil), fmt.Errorf("could not parse type: %s", fieldType.Name())
}

func parseInts(data string, count int) ([]int, error) {
	tokens := strings.SplitN(data, ",", count)

	result := make([]int, count)
	for i, s := range tokens {
		v, err := strconv.Atoi(strings.TrimSpace(s))
		if err != nil {
			return nil, err
		}
		result[i] = v
	}
	return result, nil
}

func kvp(input string) (string, string) {
	split := strings.SplitN(input, ":", 2)
	return strings.TrimSpace(split[0]), strings.TrimSpace(split[1])
}

func setSpriteDefaults(sprite *Sprite) {
	if sprite.Offsets.Offset.Eq(image.Point{}) && sprite.Offsets.OriginalSize.Eq(image.Point{}) {
		sprite.Offsets = Offsets{OriginalSize: sprite.Bounds.Size}
	}
}
