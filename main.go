package main

import (
	"crypto/aes"
	"crypto/cipher"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/sdl_image"
	"github.com/veandco/go-sdl2/sdl_ttf"
	"io/ioutil"
	"math"
	"math/rand"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	_ = iota
	up
	right
	down
	left
	PLAY
	PAUSE
	LIFE_LOST
	GAME_OVER
	LOGO
	MENU
	CREDITS
	UPGRADE
	OPTIONS
	VK_RIGHT = 1073741903
	VK_LEFT  = 1073741904
	VK_DOWN  = 1073741905
	VK_UP    = 1073741906
	VK_W     = 119
	VK_A     = 97
	VK_S     = 115
	VK_D     = 100
	VK_ESC   = 27
	VK_ENTER = 13
	VK_H     = 104
	VK_J     = 106
	VK_K     = 107
	VK_L     = 108
	win_size = 800
)

// bonus effects
const (
	_ = iota
	// good
	FREEZE
	SLOMO
	LIFE_UP
	SHIELD
	// bad
	MORE_BOMBS
	MIRROR_MODE
	MOVE_BLOCKS
	LIFE_DOWN
)

func file_exists(filename string) bool {
	if _, err := os.Stat(filename); err == nil {
		return true
	} else {
		return false
	}
}

func get_event() {
	for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
		switch e := event.(type) {
		case *sdl.WindowEvent:
			if e.Event == sdl.WINDOWEVENT_FOCUS_LOST && game.state == PLAY {
				game.state = PAUSE
			}
		case *sdl.QuitEvent:
			window.Destroy()
			os.Exit(0)
		case *sdl.KeyDownEvent:
			control.UpdateKey(int(e.Keysym.Sym), true)
		case *sdl.KeyUpEvent:
			control.UpdateKey(int(e.Keysym.Sym), false)
		}
	}
}

func random(n int) int32 {
	return int32(rand.Intn(n) + 1)
}

func random_bool() bool {
	if random(2) == 1 {
		return true
	} else {
		return false
	}
}

func start_timer() {
	started = time.Now()
}

func update_timer() {
	elapsed = time.Since(started)
}

type Storage struct {
	data           []string
	highscore, key string
	block          cipher.Block
	iv             []byte
}

func (s *Storage) init() {
	s.key = "TrYto6ue55thi5!!"
	s.block, _ = aes.NewCipher([]byte(s.key))
	ciphertext := []byte("abcdef1234567890")
	s.iv = ciphertext[:aes.BlockSize]
}

func (s *Storage) encode(str []byte) []byte {
	encrypter := cipher.NewCFBEncrypter(s.block, s.iv)
	encrypted := make([]byte, len(str))
	encrypter.XORKeyStream(encrypted, str)
	return encrypted
}

func (s *Storage) decode(str []byte) []byte {
	decrypter := cipher.NewCFBDecrypter(s.block, s.iv)
	decrypted := make([]byte, len(str))
	decrypter.XORKeyStream(decrypted, str)
	return decrypted
}

func (s *Storage) save_player() {
	str := s.highscore + "\n" +
		strconv.Itoa(int(plane.max_life)) + "\n" +
		strconv.Itoa(game.money) + "\n" +
		strconv.FormatBool(game.music) + "\n" +
		strconv.FormatBool(game.sound) + "\n" +
		strconv.Itoa(int(plane.boost))
	encoded := s.encode([]byte(str))
	ioutil.WriteFile(".player", encoded, 0644)
}

func (s *Storage) load_player() {
	bytes, _ := ioutil.ReadFile(".player")
	bytes = s.decode(bytes)
	s.data = strings.Split(string(bytes), "\n")
	s.highscore = s.data[0]
	temp, _ := strconv.Atoi(s.data[1])
	plane.max_life = int32(temp)
	game.money, _ = strconv.Atoi(s.data[2])
	game.music, _ = strconv.ParseBool(s.data[3])
	game.sound, _ = strconv.ParseBool(s.data[4])
	temp, _ = strconv.Atoi(s.data[5])
	plane.boost = int32(temp)
}

func (s *Storage) check_highscore(last int, last_str string) {
	score, _ := strconv.Atoi(strings.Replace(s.highscore, ".", "", 1))
	if last > score {
		s.highscore = last_str
		text.highscore = text.get_texture(s.highscore, 28, YELLOW)
	}
}

type Text struct {
	paused, gameover, score,
	enter, esc, cont, retry, exit,
	cont_y, retry_y, exit_y, score_value,
	level_cash, slomo, freeze, life_up, shield,
	more_bombs, mirror_mode, move_blocks,
	life_lose, play, upgrade, options, credits,
	play_y, upgrade_y, options_y, credits_y,
	game_name, your_best, highscore, code, music,
	hacked, tapetwo, mashur, oneonetwoseven,
	your_money, money, one_more_life, one_k,
	fifty, buy_r, buy_y, controls, min, max *sdl.Texture
	w, h  int32
	alpha byte
}

func (t *Text) init() {
	t.game_name = t.get_texture("?", 300, WHITE)
	t.paused = t.get_texture("paused", 48, WHITE)
	t.gameover = t.get_texture("game over", 60, WHITE)
	t.score = t.get_texture("score ", 38, WHITE)
	t.enter = t.get_texture("press ENTER to restart", 18, WHITE)
	t.esc = t.get_texture("or ESC to exit", 18, WHITE)
	t.cont = t.get_texture("continue", 28, WHITE)
	t.cont_y = t.get_texture("continue", 28, YELLOW)
	t.retry = t.get_texture("retry", 28, WHITE)
	t.retry_y = t.get_texture("retry", 28, YELLOW)
	t.exit = t.get_texture("exit", 28, WHITE)
	t.exit_y = t.get_texture("exit", 28, YELLOW)
	t.slomo = t.get_texture("slomo", 18, GREEN)
	t.freeze = t.get_texture("freeze blocks", 18, GREEN)
	t.life_up = t.get_texture("one more life", 18, GREEN)
	t.shield = t.get_texture("bomb shield", 18, GREEN)
	t.more_bombs = t.get_texture("moar bombs!", 18, RED)
	t.mirror_mode = t.get_texture("mirror mode", 18, RED)
	t.move_blocks = t.get_texture("move blocks", 18, RED)
	t.life_lose = t.get_texture("ouch", 18, RED)
	t.play = t.get_texture("play", 28, WHITE)
	t.play_y = t.get_texture("play", 28, YELLOW)
	t.upgrade = t.get_texture("upgrade", 28, WHITE)
	t.upgrade_y = t.get_texture("upgrade", 28, YELLOW)
	t.options = t.get_texture("options", 28, WHITE)
	t.options_y = t.get_texture("options", 28, YELLOW)
	t.credits = t.get_texture("credits", 28, WHITE)
	t.credits_y = t.get_texture("credits", 28, YELLOW)
	t.your_best = t.get_texture("your best:", 28, WHITE)
	t.highscore = t.get_texture(storage.highscore, 28, YELLOW)
	t.code = t.get_texture("code + art", 28, YELLOW)
	t.music = t.get_texture("music", 28, YELLOW)
	t.hacked = t.get_texture("hack3d", 22, WHITE)
	t.tapetwo = t.get_texture("tapetwo", 22, WHITE)
	t.mashur = t.get_texture("mashur", 22, WHITE)
	t.oneonetwoseven = t.get_texture("1127", 22, WHITE)
	t.your_money = t.get_texture("your cash:", 28, WHITE)
	t.one_more_life = t.get_texture("one more life:", 28, WHITE)
	t.money = t.get_texture(strconv.Itoa(game.money)+"$", 28, GREEN)
	t.one_k = t.get_texture("1000$", 28, GREEN)
	t.fifty = t.get_texture("50$", 28, GREEN)
	t.buy_r = t.get_texture("buy", 28, RED)
	t.buy_y = t.get_texture("buy", 28, YELLOW)
	t.controls = t.get_texture("controls", 28, WHITE)
	t.min = t.get_texture("min", 28, WHITE)
	t.max = t.get_texture("max", 28, WHITE)
}

func (t *Text) get_texture(text string, size int, color sdl.Color) *sdl.Texture {
	font, _ := ttf.OpenFont("JoystixInk.ttf", size)
	surf, _ := font.RenderUTF8_Blended(text, color)
	font_image, _ := renderer.CreateTextureFromSurface(surf)
	font.Close()
	return font_image
}

func (t *Text) draw_texture(texture *sdl.Texture, x, y int32, center_x, center_y bool) {
	_, _, t.w, t.h, _ = texture.Query()
	if center_x {
		x = int32(win_size/2 - t.w/2)
	}
	if center_y {
		y = int32(win_size/2 - t.h/2)
	}
	dst = sdl.Rect{x, y, int32(t.w), int32(t.h)}
	renderer.Copy(texture, &src, &dst)
}

func (t *Text) fadeout() {
	if t.alpha != 255 {
		t.alpha += 3
	}
}

type Game struct {
	loop, slomo_flag, mirror_mode, first_run, sound, music  bool
	speed, x                                                int32
	str_score                                               string
	score, saved_score, level_cash, state, money, highscore int
}

func (g *Game) check_pause() {
	if control.keys[VK_ESC] {
		if !control.esc_lock {
			switch game.state {
			case PLAY:
				pause_menu.active = 0
				g.saved_score += int(math.Trunc(elapsed.Seconds() * 100))
				g.state = PAUSE
			default:
				start_timer()
				g.state = PLAY
			}
			control.esc_lock = true
		}
	} else {
		control.esc_lock = false
	}
}

func (g *Game) game_over_event() {
	if control.keys[VK_ESC] {
		bg.alpha = 0
		text.alpha = 0
		lines.alpha = 0
		bonus.alpha = 0
		field.alpha = 0
		menu.init()
		game.state = MENU
	}
	if control.keys[VK_ENTER] {
		g.retry()
	}
}

func (g *Game) retry() {
	g.speed = 14
	g.x = 1
	plane.visible = true
	g.saved_score = 0
	g.level_cash = 0
	plane.life = plane.max_life
	plane.shielded = false
	g.state = PLAY
	g.slomo_flag = false
	g.mirror_mode = false
	bonus.spawn = false
	lines.move = false
	plane.teleport()
	start_timer()
}

func (g *Game) get_score() {
	g.score = (int(math.Trunc(elapsed.Seconds()*100)) + g.saved_score) * int(g.x)
	if g.score > 1000 {
		first := strconv.Itoa(g.score / 1000)
		second := strconv.Itoa(g.score % 1000)
		for len(second) != 3 {
			second = "0" + second
		}
		g.str_score = first + "." + second
	} else {
		g.str_score = strconv.Itoa(g.score)
	}
	storage.check_highscore(g.score, g.str_score)
	text.score_value = text.get_texture(g.str_score, 38, YELLOW)
	text.level_cash = text.get_texture("+"+strconv.Itoa(g.level_cash)+"$", 38, GREEN)
	game.money += game.level_cash
	text.money = text.get_texture(strconv.Itoa(game.money)+"$", 28, GREEN)
	storage.save_player()
}

func (g *Game) slomo() {
	if g.slomo_flag && delay < 30 {
		delay++
	}
	if !g.slomo_flag && delay > 10 {
		delay--
	}
}

type Control struct {
	keys map[int]bool
	enter_lock, up_lock,
	down_lock, left_lock,
	right_lock, esc_lock bool
}

func (c Control) UpdateKey(key int, state bool) {
	c.keys[key] = state
}

type Bg struct {
	game_bg  *sdl.Texture
	pause_bg *sdl.Texture
	alpha    byte
}

func (b *Bg) init() {
	image, _ := img.Load("pics" + slash + "bg.png")
	b.game_bg, _ = renderer.CreateTextureFromSurface(image)
	image, _ = img.Load("pics" + slash + "pause.png")
	b.pause_bg, _ = renderer.CreateTextureFromSurface(image)
	b.alpha = 0
}

func (b *Bg) draw() {
	b.game_bg.SetAlphaMod(b.alpha)
	dst = sdl.Rect{0, 0, win_size, win_size}
	renderer.Copy(b.game_bg, &src, &dst)
}

func (b *Bg) draw_pause() {
	dst = sdl.Rect{0, 250, win_size, 300}
	renderer.Copy(bg.pause_bg, &src, &dst)
}

func (b *Bg) fadeout() {
	if b.alpha != 255 {
		b.alpha += 3
	}
}

type Block struct {
	rect             sdl.Rect
	direction, blood bool
}

type Lines struct {
	pic, broken_pic *sdl.Texture
	line            [5][]Block
	move            bool
	alpha           byte
}

func (l *Lines) init() {
	image, _ := img.Load("pics" + slash + "block.png")
	l.pic, _ = renderer.CreateTextureFromSurface(image)
	image, _ = img.Load("pics" + slash + "broken_block.png")
	l.broken_pic, _ = renderer.CreateTextureFromSurface(image)
	image = nil
	l.move = false
	for i := 0; i < 5; i++ {
		l.line[i] = []Block{}
	}
}

func (l *Lines) move_block(block *Block) {
	switch plane.direction {
	case left, right:
		if block.direction { // up
			if block.rect.Y != 0 {
				block.rect.Y--
			} else {
				block.direction = !block.direction
			}
		} else { // down
			if block.rect.Y < win_size {
				block.rect.Y++
			} else {
				block.direction = !block.direction
			}
		}
		break
	case up, down:
		if block.direction { // left
			if block.rect.X != 0 {
				block.rect.X--
			} else {
				block.direction = !block.direction
			}
		} else { // right
			if block.rect.X < win_size {
				block.rect.X++
			} else {
				block.direction = !block.direction
			}
		}
		break
	}
}

func (l *Lines) move_lines() {
	if l.move {
		for i := 0; i < len(l.line[plane.direction]); i++ {
			l.move_block(&l.line[plane.direction][i])
		}
	}
}

func (l *Lines) draw() {
	l.pic.SetAlphaMod(l.alpha)
	for i := 0; i < len(l.line[plane.direction]); i++ {
		dst = sdl.Rect{l.line[plane.direction][i].rect.X, l.line[plane.direction][i].rect.Y, 25, 25}
		if l.line[plane.direction][i].blood {
			renderer.Copy(l.broken_pic, &src, &dst)
		} else {
			renderer.Copy(l.pic, &src, &dst)
		}

	}
}

func (l *Lines) create_blocks(count int32) {
	switch plane.direction {
	case right:
		for i := 0; i < int(count); i++ {
			l.line[plane.direction] = append(l.line[plane.direction], Block{sdl.Rect{775, random(31) * 25, 25, 25}, random_bool(), false})
		}
		break
	case left:
		for i := 0; i < int(count); i++ {
			l.line[plane.direction] = append(l.line[plane.direction], Block{sdl.Rect{0, random(31) * 25, 25, 25}, random_bool(), false})
		}
		break
	case up:
		for i := 0; i < int(count); i++ {
			l.line[plane.direction] = append(l.line[plane.direction], Block{sdl.Rect{random(31) * 25, 0, 25, 25}, random_bool(), false})
		}
		break
	case down:
		for i := 0; i < int(count); i++ {
			l.line[plane.direction] = append(l.line[plane.direction], Block{sdl.Rect{random(31) * 25, 775, 25, 25}, random_bool(), false})
		}
		break
	}
}

func (l *Lines) clear_line() {
	l.line[plane.direction] = []Block{}
}

func (l *Lines) fadeout() {
	if l.alpha != 255 {
		l.alpha += 5
	}
}

type Plane struct {
	x, y, speed, direction, boost, life, max_life int32
	hp, no_hp, shield, image                      *sdl.Texture
	rect                                          *sdl.Rect
	shielded                                      bool
	angle                                         float64
	visible                                       bool
}

func (p *Plane) init() {
	p.speed = 0
	p.direction = 1
	if p.boost == 0 {
		p.boost = 10
	}
	p.teleport()
	if p.max_life == 0 {
		p.max_life = 1
	}
	p.life = p.max_life
	p.visible = true
	image, _ := img.Load("pics" + slash + "hp.png")
	p.hp, _ = renderer.CreateTextureFromSurface(image)
	image, _ = img.Load("pics" + slash + "no_hp.png")
	p.no_hp, _ = renderer.CreateTextureFromSurface(image)
	image, _ = img.Load("pics" + slash + "shield.png")
	p.shield, _ = renderer.CreateTextureFromSurface(image)
	image, _ = img.Load("pics" + slash + "plane.png")
	p.image, _ = renderer.CreateTextureFromSurface(image)
}

func (p *Plane) borders() {
	switch p.direction {
	case right, left:
		if p.y-p.speed >= 0 && p.y-p.speed <= 781 {
			p.y -= p.speed
		}
		if p.direction == right {
			p.x += game.speed / 5
		} else {
			p.x -= game.speed / 5
		}
		break
	case down, up:
		if p.x-p.speed >= 0 && p.x-p.speed <= 781 {
			p.x -= p.speed
		}
		if p.direction == down {
			p.y += game.speed / 5
		} else {
			p.y -= game.speed / 5
		}
		break
	}
}

func (p *Plane) move() {
	var steer_key1, steer_key2 bool
	switch p.direction {
	case left, right:
		steer_key1 = control.keys[VK_UP] || control.keys[VK_W] || control.keys[VK_K]
		steer_key2 = control.keys[VK_DOWN] || control.keys[VK_S] || control.keys[VK_J]
		break
	case up, down:
		steer_key1 = control.keys[VK_LEFT] || control.keys[VK_A] || control.keys[VK_H]
		steer_key2 = control.keys[VK_RIGHT] || control.keys[VK_D] || control.keys[VK_L]
		break
	}
	if game.mirror_mode {
		steer_key1, steer_key2 = steer_key2, steer_key1
	}
	if steer_key1 == false {
		if p.speed > 0 {
			p.speed--
		}
	}
	if steer_key2 == false {
		if p.speed < 0 {
			p.speed++
		}
	}
	if steer_key1 {
		if p.speed < p.boost && p.speed >= 0 {
			p.speed++
			game.saved_score++
		}
	}
	if steer_key2 {
		if p.speed > -p.boost && p.speed <= 0 {
			p.speed--
			game.saved_score++
		}
	}
	p.borders()
}

func (p *Plane) draw() {
	if p.visible {
		dst = sdl.Rect{p.x, p.y, 19, 19}
		if game.mirror_mode {
			renderer.CopyEx(p.image, &src, &dst, p.angle+180, nil, sdl.FLIP_NONE)
		} else {
			renderer.CopyEx(p.image, &src, &dst, p.angle, nil, sdl.FLIP_NONE)
		}
	}
	if p.shielded {
		dst = sdl.Rect{p.x - 2, p.y - 2, 23, 23}
		renderer.CopyEx(p.shield, &src, &dst, p.angle, nil, sdl.FLIP_NONE)
	}
}

func (p *Plane) teleport() {
	if bonus.show_text {
		bonus.show_text = false
	}
	go bonus.counter()
	go lines.clear_line()
	go field.clear()
	p.direction = random(4)
	p.angle = float64((p.direction - 1) * 90)
	if p.direction%2 == 0 {
		p.y = random(740)
		if p.direction == right {
			p.x = -150
		} else {
			p.x = 950
		}
	} else {
		p.x = random(740)
		if p.direction == down {
			p.y = -150
		} else {
			p.y = 950
		}
	}
	if game.x != 7 {
		game.speed++
		game.x = game.speed/5 - 2
	}
	go field.create_bombs(int(random(int(game.x+2) / 2)))
	if !bonus.active {
		go bonus.randomize()
	}
	lines.move = random_bool()
	if lines.move {
		go lines.create_blocks(random(int(game.x)))
	} else {
		go lines.create_blocks(game.x + random(int((game.x+1)/2)))
	}
}

func (p *Plane) out() bool {
	switch p.direction {
	case up:
		if p.y < 0 {
			return true
		} else {
			return false
		}
		break
	case right:
		if p.x > win_size {
			return true
		} else {
			return false
		}
		break
	case down:
		if p.y > win_size {
			return true
		} else {
			return false
		}
		break
	case left:
		if p.x < 0 {
			return true
		} else {
			return false
		}
		break
	}
	return false
}

func (p *Plane) death() {
	if p.life-1 <= 0 {
		p.life--
		game.get_score()
		if game.first_run {
			game.first_run = false
		}
		bonus.active = false
		bonus.count = 0
		game.slomo_flag = false
		game.mirror_mode = false
		delay = 10
		game.state = GAME_OVER
	} else {
		game.saved_score += int(math.Trunc(elapsed.Seconds() * 100))
		start_timer()
		game.state = LIFE_LOST
	}
}

func (p *Plane) draw_life() {
	for i := 0; i < 3; i++ {
		dst = sdl.Rect{int32(365 + 25*i), 365, 18, 18}
		renderer.Copy(p.no_hp, &src, &dst)
	}
	for i := 0; i < int(p.life); i++ {
		dst = sdl.Rect{int32(365 + 25*i), 365, 18, 18}
		renderer.Copy(p.hp, &src, &dst)
	}
}

func (p *Plane) collide_block(block *Block) {
	if p.rect.HasIntersection(&block.rect) {
		block.blood = true
		p.death()
	}
}

func (p *Plane) collide_blocks() {
	p.rect = &dst
	for i := 0; i < len(lines.line[plane.direction]); i++ {
		p.collide_block(&lines.line[plane.direction][i])
	}
}

func (p *Plane) collide_field() {
	p.rect = &dst
	for i := 0; i < len(field.list); i++ {
		if p.rect.HasIntersection(&field.list[i].rect) {
			field.coll_rect = field.list[i].rect
			field.animate = true
			field.remove(i)
			if p.shielded {
				p.shielded = false
			} else {
				p.visible = false
				p.death()
			}
		}
	}
}

func (p *Plane) collide_bonus() {
	p.rect = &dst
	if bonus.spawn && p.rect.HasIntersection(&bonus.rect) {
		bonus.spawn = false
		bonus.apply_effect()
	}
}

func (p *Plane) check_collision() {
	switch p.direction {
	case right:
		if p.x > 770 {
			p.collide_blocks()
		}
		if p.x > 50 {
			p.collide_field()
			p.collide_bonus()
		}
		break
	case left:
		if p.x < 30 {
			p.collide_blocks()
		}
		if p.x < 750 {
			p.collide_field()
			p.collide_bonus()
		}
		break
	case up:
		if p.y < 30 {
			p.collide_blocks()
		}
		if p.y < 750 {
			p.collide_field()
			p.collide_bonus()
		}
		break
	case down:
		if p.y > 770 {
			p.collide_blocks()
		}
		if p.y > 50 {
			p.collide_field()
			p.collide_bonus()
		}
		break
	}
}

type Bomb struct {
	rect       sdl.Rect
	tick, m, n int
	blink      bool
}

func (b *Bomb) add_tick() {
	if b.tick < 50 {
		b.tick++
	} else {
		b.tick = 0
		b.blink = !b.blink
	}
}

type Bonus struct {
	rect                                     sdl.Rect
	pic                                      *sdl.Texture
	effect, count, tp_counter, active_effect int32
	spawn, show_text, active                 bool
	alpha                                    byte
}

func (b *Bonus) init() {
	image, _ := img.Load("pics" + slash + "random.png")
	b.pic, _ = renderer.CreateTextureFromSurface(image)
	b.rect = sdl.Rect{0, 0, 22, 22}
	b.spawn = false
	b.count = 0
	b.tp_counter = 0
	b.show_text = false
	b.active = false
}

func (b *Bonus) draw() {
	if b.spawn {
		dst = b.rect
		b.pic.SetAlphaMod(b.alpha)
		renderer.Copy(b.pic, &src, &dst)
	}
}

func (b *Bonus) text() {
	if b.show_text {
		var texture *sdl.Texture
		switch b.effect {
		case SLOMO:
			texture = text.slomo
		case FREEZE:
			texture = text.freeze
		case LIFE_UP:
			texture = text.life_up
		case SHIELD:
			texture = text.shield
		case MORE_BOMBS:
			texture = text.more_bombs
		case MIRROR_MODE:
			texture = text.mirror_mode
		case MOVE_BLOCKS:
			texture = text.move_blocks
		case LIFE_DOWN:
			texture = text.life_lose
		}
		text.draw_texture(texture, b.rect.X, b.rect.Y-20, false, false)
	}
}

func (b *Bonus) randomize() {
	var good bool
	b.spawn = random_bool()
	if b.spawn {
		mn_group := field.free_cell()
		b.rect.X = int32(150 + mn_group[0]*30 + 4)
		b.rect.Y = int32(150 + mn_group[1]*30 + 4)
		good = random_bool()
		if good {
			b.effect = random(4)
		} else {
			b.effect = 4 + random(4)
		}
	}
}

func (b *Bonus) counter() {
	if b.active && b.count < 3 {
		b.count++
	}
	if b.count == 3 {
		b.active = false
		b.count = 0
		switch b.active_effect {
		case SLOMO:
			game.saved_score += int(math.Trunc(elapsed.Seconds()*100)) / 5
			game.slomo_flag = false
		case MIRROR_MODE:
			game.mirror_mode = false
		}
	}
}

func (b *Bonus) apply_effect() {
	b.show_text = true
	switch b.effect {
	case SLOMO:
		game.slomo_flag = true
		b.active = true
		b.active_effect = SLOMO
		game.saved_score += int(math.Trunc(elapsed.Seconds() * 100))
		start_timer()
	case FREEZE:
		lines.move = false
	case LIFE_UP:
		if plane.life != 3 {
			plane.life++
		}
	case SHIELD:
		plane.shielded = true
	case MORE_BOMBS:
		field.create_bombs(3)
	case MIRROR_MODE:
		game.mirror_mode = true
		b.active_effect = MIRROR_MODE
		b.active = true
	case MOVE_BLOCKS:
		lines.move = true
	case LIFE_DOWN:
		plane.death()
	}
}

func (b *Bonus) fadeout() {
	if b.alpha != 255 {
		b.alpha += 3
	}
}

type Field struct {
	list                 []Bomb
	pic, pic_blink, boom *sdl.Texture
	angle                float64
	boom_rect            *sdl.Rect
	coll_rect            sdl.Rect
	animate              bool
	tick                 int
	alpha                byte
}

func (f *Field) init() {
	image, _ := img.Load("pics" + slash + "mine.png")
	f.pic, _ = renderer.CreateTextureFromSurface(image)
	image, _ = img.Load("pics" + slash + "mine_blink.png")
	f.pic_blink, _ = renderer.CreateTextureFromSurface(image)
	image, _ = img.Load("pics" + slash + "explosion.png")
	f.boom, _ = renderer.CreateTextureFromSurface(image)
	f.angle = 0
	f.boom_rect = &sdl.Rect{0, 0, 64, 64}
}

func (f *Field) remove(num int) {
	f.list = append(f.list[:num], f.list[num+1:]...)
}

func (f *Field) clear() {
	f.list = []Bomb{}
}

func (f *Field) free_cell() [2]int {
	var m, n int32
search:
	m = random(15)
	n = random(15)
	for i := 0; i < len(f.list); i++ {
		if int(m) == f.list[i].m && int(n) == f.list[i].n {
			goto search // i'm really sorry :(
		}
	}
	return [2]int{int(m), int(n)}
}

func (f *Field) create_bombs(amount int) {
	for amount != 0 {
		mn_group := f.free_cell()
		f.list = append(f.list, Bomb{sdl.Rect{int32(150 + mn_group[0]*30), int32(150 + mn_group[1]*30), 30, 30}, int(random(50)), mn_group[0], mn_group[1], random_bool()})
		amount--
	}
}

func (f *Field) draw() {
	if f.animate {
		f.explosion()
		dst = sdl.Rect{f.coll_rect.X, f.coll_rect.Y, 64, 64}
		renderer.Copy(f.boom, f.boom_rect, &dst)
	}
	for _, i := range f.list {
		dst = i.rect
		f.pic.SetAlphaMod(f.alpha)
		f.pic_blink.SetAlphaMod(f.alpha)
		if i.blink {
			renderer.CopyEx(f.pic_blink, &src, &dst, f.angle, nil, sdl.FLIP_NONE)
		} else {
			renderer.CopyEx(f.pic, &src, &dst, f.angle, nil, sdl.FLIP_NONE)
		}
	}
}

func (f *Field) add_tick_bombs() {
	for i := 0; i < len(f.list); i++ {
		f.list[i].add_tick()
	}
}

func (f *Field) rotate_bombs() {
	f.angle += 5
}

func (f *Field) explosion() {
	if f.animate {
		if f.tick < 32 {
			f.boom_rect.X = int32(f.tick / 2 % 4 * 64)
			f.boom_rect.Y = int32(f.tick / 8 * 64)
			f.tick++
		} else {
			f.animate = false
			f.tick = 0
			f.coll_rect.X = 0
			f.coll_rect.Y = 0
		}
	}
}

func (f *Field) fadeout() {
	if f.alpha != 255 {
		f.alpha += 3
	}
}

type Logo struct {
	left_pic, right_pic, hacked             *sdl.Texture
	alpha                                   byte
	left_x, left_y, right_x, right_y, pause int32
	fade                                    bool
}

func (l *Logo) init() {
	image, _ := img.Load("pics" + slash + "left_logo.png")
	l.left_pic, _ = renderer.CreateTextureFromSurface(image)
	image, _ = img.Load("pics" + slash + "right_logo.png")
	l.right_pic, _ = renderer.CreateTextureFromSurface(image)
	image, _ = img.Load("pics" + slash + "hacked.png")
	l.hacked, _ = renderer.CreateTextureFromSurface(image)
	l.alpha = 0
	l.pause = 0
	l.left_x = 295
	l.left_y = 333
	l.right_x = 403
	l.right_y = 333
}

func (l *Logo) draw() {
	if l.alpha == 255 || l.left_x == 179 {
		dst = sdl.Rect{300, 379, 210, 43}
		renderer.Copy(l.hacked, &src, &dst)
	}
	dst = sdl.Rect{l.left_x, l.left_y, 109, 145}
	renderer.Copy(l.left_pic, &src, &dst)
	dst = sdl.Rect{l.right_x, l.right_y, 107, 145}
	renderer.Copy(l.right_pic, &src, &dst)
}

func (l *Logo) update() {
	if control.keys[VK_ESC] || control.keys[VK_ENTER] {
		game.state = MENU
		return
	}
	if l.alpha != 255 && !l.fade {
		l.alpha += 3
		l.left_pic.SetAlphaMod(l.alpha)
		l.right_pic.SetAlphaMod(l.alpha)
	} else {
		l.fade = true
		if l.left_y != 295 {
			l.left_y--
		}
		if l.right_y != 368 {
			l.right_y++
		}
		if l.left_y == 295 && l.right_y == 368 {
			if l.left_x != 179 {
				l.left_x -= 2
			}
			if l.right_x != 521 {
				l.right_x += 2
			}
			if l.left_x == 179 && l.right_x == 521 && l.fade {
				l.pause++
				if l.pause > 100 {
					l.alpha -= 3
					l.left_pic.SetAlphaMod(l.alpha)
					l.right_pic.SetAlphaMod(l.alpha)
					l.hacked.SetAlphaMod(l.alpha)
					if l.alpha == 0 {
						game.state = MENU
					}
				}
			}
		}
	}
}

type Pause_menu struct {
	list               [3]string
	active             int
	up_lock, down_lock bool
}

func (p *Pause_menu) init() {
	p.list = [3]string{"CONTINUE", "RETRY", "EXIT"}
	p.active = 0
}

func (p *Pause_menu) events() {
	if control.keys[VK_DOWN] || control.keys[VK_S] || control.keys[VK_J] {
		if !control.down_lock && p.active < 2 {
			p.active++
			control.down_lock = true
		}
	} else {
		control.down_lock = false
	}
	if control.keys[VK_UP] || control.keys[VK_W] || control.keys[VK_K] {
		if !control.up_lock && p.active > 0 {
			p.active--
			control.up_lock = true
		}
	} else {
		control.up_lock = false
	}
	if control.keys[VK_ENTER] {
		control.enter_lock = true
		switch p.list[p.active] {
		case "CONTINUE":
			game.state = PLAY
			start_timer()
		case "RETRY":
			game.retry()
			game.state = PLAY
		case "EXIT":
			game.slomo_flag = false
			bg.alpha = 0
			text.alpha = 0
			lines.alpha = 0
			bonus.alpha = 0
			field.alpha = 0
			menu.init()
			game.state = MENU
		}
	} else {
		control.enter_lock = false
	}
}

func (p *Pause_menu) draw(bg_only bool) {
	bg.draw_pause()
	if !bg_only {
		plane.draw_life()
		text.draw_texture(text.paused, 0, 280, true, false)
		for num := range p.list {
			if num == p.active {
				switch num {
				case 0:
					text.draw_texture(text.cont_y, 0, int32(420+num*30), true, false)
				case 1:
					text.draw_texture(text.retry_y, 0, int32(420+num*30), true, false)
				case 2:
					text.draw_texture(text.exit_y, 0, int32(420+num*30), true, false)
				}
			} else {
				switch num {
				case 0:
					text.draw_texture(text.cont, 0, int32(420+num*30), true, false)
				case 1:
					text.draw_texture(text.retry, 0, int32(420+num*30), true, false)
				case 2:
					text.draw_texture(text.exit, 0, int32(420+num*30), true, false)
				}
			}
		}
	}
}

type Menu struct {
	list           [5]string
	active, last   int
	sum, width_sum int32
	offset         [5]int32
	width          [5]int32
	text_width     [2]int32
}

func (m *Menu) init() {
	m.list = [5]string{"PLAY", "UPGRADE", "OPTIONS", "CREDITS", "EXIT"}
	m.last = -1
	m.offset = [5]int32{800, 750, 700, 650, 600}
	_, _, m.text_width[0], _, _ = text.your_best.Query()
	_, _, m.text_width[1], _, _ = text.highscore.Query()
	m.width_sum = 0
	game.slomo_flag = false
	delay = 10
	for _, value := range m.text_width {
		m.width_sum += int32(value)
	}
	m.width_sum = 400 - m.width_sum/2
}

func (m *Menu) events() {
	m.sum = 0
	for _, value := range m.offset {
		m.sum += value
	}
	if m.sum <= 0 {
		if control.keys[VK_DOWN] || control.keys[VK_S] || control.keys[VK_J] {
			if !control.down_lock && m.active < 4 {
				m.last = m.active
				m.active++
				control.down_lock = true
			}
		} else {
			control.down_lock = false
		}
		if control.keys[VK_UP] || control.keys[VK_W] || control.keys[VK_K] {
			if !control.up_lock && m.active > 0 {
				m.last = m.active
				m.active--
				control.up_lock = true
			}
		} else {
			control.up_lock = false
		}
		if control.keys[VK_ENTER] {
			if !control.enter_lock {
				control.enter_lock = true
				switch m.list[m.active] {
				case "PLAY":
					game.retry()
					game.state = PLAY
					start_timer()
				case "CREDITS":
					game.state = CREDITS
				case "UPGRADE":
					if plane.max_life != 3 {
						upgrade.init()
						game.state = UPGRADE
					}
				case "OPTIONS":
					game.state = OPTIONS
				case "EXIT":
					window.Destroy()
					os.Exit(0)
				}
			}
		} else {
			control.enter_lock = false
		}
	}
}

func (m *Menu) draw() {
	bg.draw()
	text.game_name.SetAlphaMod(text.alpha)
	text.draw_texture(text.game_name, 0, 100, true, false)
	if !game.first_run {
		text.your_best.SetAlphaMod(text.alpha)
		text.highscore.SetAlphaMod(text.alpha)
		text.draw_texture(text.your_best, m.width_sum, 400, false, false)
		text.draw_texture(text.highscore, m.width_sum+190, 400, false, false)
	}
	for num := range m.list {
		if num == m.active {
			switch num {
			case 0:
				_, _, text.w, text.h, _ = text.play_y.Query()
				text.draw_texture(text.play_y, int32(790-text.w)+m.offset[num], int32(640+num*30), false, false)
			case 1:
				_, _, text.w, text.h, _ = text.upgrade_y.Query()
				text.draw_texture(text.upgrade_y, int32(790-text.w)+m.offset[num], int32(640+num*30), false, false)
			case 2:
				_, _, text.w, text.h, _ = text.options_y.Query()
				text.draw_texture(text.options_y, int32(790-text.w)+m.offset[num], int32(640+num*30), false, false)
			case 3:
				_, _, text.w, text.h, _ = text.credits_y.Query()
				text.draw_texture(text.credits_y, int32(790-text.w)+m.offset[num], int32(640+num*30), false, false)
			case 4:
				_, _, text.w, text.h, _ = text.exit_y.Query()
				text.draw_texture(text.exit_y, int32(790-text.w)+m.offset[num], int32(640+num*30), false, false)
			}
		} else {
			switch num {
			case 0:
				_, _, text.w, text.h, _ = text.play.Query()
				text.draw_texture(text.play, int32(790-text.w)+m.offset[num], int32(640+num*30), false, false)
			case 1:
				_, _, text.w, text.h, _ = text.upgrade.Query()
				text.draw_texture(text.upgrade, int32(790-text.w)+m.offset[num], int32(640+num*30), false, false)
			case 2:
				_, _, text.w, text.h, _ = text.options.Query()
				text.draw_texture(text.options, int32(790-text.w)+m.offset[num], int32(640+num*30), false, false)
			case 3:
				_, _, text.w, text.h, _ = text.credits.Query()
				text.draw_texture(text.credits, int32(790-text.w)+m.offset[num], int32(640+num*30), false, false)
			case 4:
				_, _, text.w, text.h, _ = text.exit.Query()
				text.draw_texture(text.exit, int32(790-text.w)+m.offset[num], int32(640+num*30), false, false)
			}
		}
	}
}

func (m *Menu) update() {
	for i := 0; i < 5; i++ {
		if m.offset[i] > 0 {
			m.offset[i] -= 10
		}
	}
	m.slide()
}

func (m *Menu) slide() {
	if m.sum <= 0 {
		if m.last != -1 {
			if m.offset[m.last] != 0 {
				m.offset[m.last] += 3
			}
		}
		if m.offset[m.active] > -30 {
			m.offset[m.active] -= 3
		}
	}
}

type Credits struct{}

func (c *Credits) draw() {
	text.draw_texture(text.code, 0, 270, true, false)
	text.draw_texture(text.hacked, 0, 320, true, false)
	text.draw_texture(text.music, 0, 370, true, false)
	text.draw_texture(text.oneonetwoseven, 0, 415, true, false)
	text.draw_texture(text.tapetwo, 0, 455, true, false)
	text.draw_texture(text.mashur, 0, 495, true, false)
}

func (c *Credits) events() {
	if control.keys[VK_ESC] {
		game.state = MENU
	}
}

type Upgrade struct {
	width, width2         [2]int32
	width_sum, width_sum2 int32
	cost                  int
}

func (u *Upgrade) init() {
	plane.life = plane.max_life
	_, _, u.width[0], _, _ = text.your_money.Query()
	_, _, u.width[1], _, _ = text.money.Query()
	u.width_sum = 0
	for _, value := range u.width {
		u.width_sum += int32(value)
	}
	u.width_sum = 400 - u.width_sum/2
	_, _, u.width2[0], _, _ = text.one_more_life.Query()
	switch plane.max_life {
	case 1:
		_, _, u.width2[1], _, _ = text.fifty.Query()
	case 2:
		_, _, u.width2[1], _, _ = text.one_k.Query()
	}
	u.width_sum2 = 0
	for _, value := range u.width2 {
		u.width_sum2 += int32(value)
	}
	u.width_sum2 = 400 - u.width_sum/2
	if plane.max_life == 1 {
		u.cost = 50
	} else {
		u.cost = 1000
	}
}

func (u *Upgrade) draw() {
	pause_menu.draw(true)
	plane.draw_life()
	text.draw_texture(text.your_money, u.width_sum, 295, false, false)
	text.draw_texture(text.money, u.width_sum+200, 295, false, false)
	switch plane.max_life {
	case 1:
		text.draw_texture(text.one_more_life, u.width_sum2-40, 420, false, false)
		text.draw_texture(text.fifty, u.width_sum2+230, 420, false, false)
	case 2:
		text.draw_texture(text.one_more_life, u.width_sum2-40, 420, false, false)
		text.draw_texture(text.one_k, u.width_sum2+230, 420, false, false)
	}
	if game.money < u.cost {
		text.draw_texture(text.buy_r, 0, 480, true, false)
	} else {
		text.draw_texture(text.buy_y, 0, 480, true, false)
	}
}

func (u *Upgrade) events() {
	if control.keys[VK_ESC] {
		game.state = MENU
	}
	if control.keys[VK_ENTER] {
		if !control.enter_lock {
			control.enter_lock = true
			if game.money >= u.cost && plane.max_life < 3 {
				game.money -= u.cost
				plane.life++
				plane.max_life++
				text.money = text.get_texture(strconv.Itoa(game.money)+"$", 28, GREEN)
				u.init()
				storage.save_player()
			}
		}
	} else {
		control.enter_lock = false
	}
}

type Options struct {
	active            int
	sound, music, off *sdl.Texture
}

func (o *Options) init() {
	image, _ := img.Load("pics" + slash + "music.png")
	o.music, _ = renderer.CreateTextureFromSurface(image)
	image, _ = img.Load("pics" + slash + "sound.png")
	o.sound, _ = renderer.CreateTextureFromSurface(image)
	image, _ = img.Load("pics" + slash + "off.png")
	o.off, _ = renderer.CreateTextureFromSurface(image)
}

func (o *Options) draw() {
	pause_menu.draw(true)
	text.draw_texture(text.controls, 0, 400, true, false)
	text.draw_texture(text.min, 35, 467, false, false)
	text.draw_texture(text.max, 705, 467, false, false)
	renderer.Copy(o.music, &src, &sdl.Rect{130, 290, 89, 89})
	renderer.Copy(o.sound, &src, &sdl.Rect{620, 300, 64, 64})
	if !game.music {
		renderer.Copy(o.off, &src, &sdl.Rect{145, 305, 65, 64})
	}
	if !game.sound {
		renderer.Copy(o.off, &src, &sdl.Rect{620, 300, 65, 64})
	}
	renderer.SetDrawColor(255, 255, 255, 255)
	renderer.FillRect(&sdl.Rect{125, 480, 550, 7})
	if o.active == 2 {
		renderer.SetDrawColor(255, 255, 65, 255)
	}
	renderer.FillRect(&sdl.Rect{125 + (plane.boost-5)*55, 470, 7, 28})
	switch o.active {
	case 0:
		renderer.SetDrawColor(255, 255, 65, 255)
		renderer.FillRect(&sdl.Rect{130, 390, 90, 7})
	case 1:
		renderer.SetDrawColor(255, 255, 65, 255)
		renderer.FillRect(&sdl.Rect{605, 390, 90, 7})
	}
	renderer.SetDrawColor(0, 0, 0, 255)
}

func (o *Options) events() {
	if control.keys[VK_ESC] {
		o.active = 0
		storage.save_player()
		game.state = MENU
	}
	if control.keys[VK_ENTER] {
		if !control.enter_lock {
			control.enter_lock = true
			switch o.active {
			case 0:
				game.music = !game.music
			case 1:
				game.sound = !game.sound
			}
		}
	} else {
		control.enter_lock = false
	}
	if control.keys[VK_DOWN] || control.keys[VK_S] || control.keys[VK_J] {
		if !control.down_lock {
			control.down_lock = true
			o.active = 2
		}
	} else {
		control.down_lock = false
	}
	if control.keys[VK_UP] || control.keys[VK_W] || control.keys[VK_K] {
		if !control.up_lock {
			control.up_lock = true
			o.active = 0
		}
	} else {
		control.up_lock = false
	}
	if control.keys[VK_LEFT] || control.keys[VK_A] || control.keys[VK_H] {
		if !control.left_lock {
			control.left_lock = true
			switch o.active {
			case 2:
				if plane.boost != 5 {
					plane.boost--
				}
			default:
				o.active = 0
			}
		}
	} else {
		control.left_lock = false
	}
	if control.keys[VK_RIGHT] || control.keys[VK_D] || control.keys[VK_L] {
		if !control.right_lock {
			control.right_lock = true
			switch o.active {
			case 2:
				if plane.boost != 15 {
					plane.boost++
				}
			default:
				o.active = 1
			}

		}
	} else {
		control.right_lock = false
	}
}

var (
	bg           = Bg{}
	plane        = Plane{}
	control      = Control{map[int]bool{}, false, false, false, false, false, false}
	game         = Game{true, false, false, false, true, true, 14, 1, "", 0, 0, 0, LOGO, 0, 0}
	window       *sdl.Window
	renderer     *sdl.Renderer
	src          = sdl.Rect{0, 0, win_size, win_size}
	dst          = sdl.Rect{}
	lines        = Lines{}
	started      = time.Time{}
	elapsed      = time.Duration(1)
	slash        string
	delay        uint32 = 10
	WHITE               = sdl.Color{255, 255, 255, 255}
	YELLOW              = sdl.Color{255, 255, 65, 255}
	GREEN               = sdl.Color{60, 200, 0, 255}
	RED                 = sdl.Color{255, 0, 0, 255}
	pause_menu          = Pause_menu{}
	text                = Text{}
	interval     int    = 0
	life_removed bool   = false
	field               = Field{}
	bonus               = Bonus{}
	logo                = Logo{}
	menu                = Menu{}
	storage             = Storage{}
	credits             = Credits{}
	upgrade             = Upgrade{}
	options             = Options{}
)

func init() {
	window, _ = sdl.CreateWindow("?", sdl.WINDOWPOS_CENTERED,
		sdl.WINDOWPOS_CENTERED, win_size, win_size,
		sdl.WINDOW_SHOWN)
	renderer, _ = sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)

	cpu := runtime.NumCPU()
	runtime.GOMAXPROCS(cpu)

	switch runtime.GOOS {
	case "linux":
		slash = "/"
	case "windows":
		slash = "\\"
	}

	storage.init()
	if file_exists(".player") {
		storage.load_player()
	} else {
		game.first_run = true
		storage.highscore = "0"
	}
	rand.Seed(time.Now().UnixNano())
	ttf.Init()
	lines.init()
	plane.init()
	bg.init()
	pause_menu.init()
	text.init()
	field.init()
	bonus.init()
	logo.init()
	options.init()
	menu.init()
}

func draw_world() {
	bg.draw()
	lines.draw()
	field.draw()
	bonus.draw()
	bonus.text()
	plane.draw()
}

func render() {
	switch game.state {
	case LOGO:
		logo.draw()
	case MENU:
		menu.draw()
	case CREDITS:
		menu.draw()
		pause_menu.draw(true)
		credits.draw()
	case UPGRADE:
		menu.draw()
		upgrade.draw()
	case OPTIONS:
		menu.draw()
		options.draw()
	case PLAY:
		draw_world()
		break
	case PAUSE:
		draw_world()
		pause_menu.draw(false)
		break
	case LIFE_LOST:
		draw_world()
		pause_menu.draw(true)
		plane.draw_life()
		break
	case GAME_OVER:
		draw_world()
		text.draw_texture(text.gameover, 0, 0, true, true)
		text.draw_texture(text.score, 250, 430, false, false)
		text.draw_texture(text.level_cash, 0, 480, true, false)
		text.draw_texture(text.score_value, 410, 430, false, false)
		text.draw_texture(text.enter, 0, 630, true, false)
		text.draw_texture(text.esc, 0, 660, true, false)
		break
	}
}

func logic() {
	switch game.state {
	case LOGO:
		logo.update()
	case MENU:
		go menu.events()
		go bg.fadeout()
		go text.fadeout()
		go menu.update()
	case CREDITS:
		credits.events()
	case UPGRADE:
		upgrade.events()
	case OPTIONS:
		options.events()
	case PLAY:
		game.check_pause()
		bonus.fadeout()
		field.fadeout()
		lines.fadeout()
		plane.move()
		lines.move_lines()
		field.add_tick_bombs()
		field.rotate_bombs()
		plane.check_collision()
		game.slomo()
		if plane.out() {
			game.level_cash++
			plane.teleport()

		}
		update_timer()
		break
	case PAUSE:
		game.check_pause()
		pause_menu.events()
		break
	case LIFE_LOST:
		update_timer()
		interval = int(math.Trunc(elapsed.Seconds() * 100))
		if interval >= 50 && !life_removed {
			start_timer()
			plane.life--
			life_removed = true
			interval = 0
			break
		}
		if interval >= 50 && life_removed {
			life_removed = false
			game.state = PLAY
			plane.visible = true
			plane.teleport()
			start_timer()
			break
		}
	case GAME_OVER:
		game.game_over_event()
		break
	}
}

func main() {
	for game.loop {
		get_event()
		logic()
		renderer.Clear()
		render()
		renderer.Present()
		sdl.Delay(delay)
	}
}
