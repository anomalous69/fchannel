package routes

import (
	"strconv"

	"github.com/FChannel0/FChannel-Server/activitypub"
	"github.com/FChannel0/FChannel-Server/config"
	"github.com/FChannel0/FChannel-Server/post"
	"github.com/FChannel0/FChannel-Server/util"
	"github.com/FChannel0/FChannel-Server/webfinger"
	"github.com/gofiber/fiber/v2"
)

func Outbox(ctx *fiber.Ctx) error {
	actor, err := webfinger.GetActorFromPath(ctx.Path(), "/")
	if err != nil {
		return util.MakeError(err, "Outbox")
	}

	if activitypub.AcceptActivity(ctx.Get("Accept")) {
		actor.GetOutbox(ctx)
		return nil
	}

	return ParseOutboxRequest(ctx, actor)
}

func OutboxGet(ctx *fiber.Ctx) error {
	actor, err := activitypub.GetActorByNameFromDB(ctx.Params("actor"))

	if err != nil {
		return nil
	}

	if activitypub.AcceptActivity(ctx.Get("Accept")) {
		actor.GetInfoResp(ctx)
		return nil
	}

	var page int
	if postNum := ctx.Query("page"); postNum != "" {
		if page, err = strconv.Atoi(postNum); err != nil {
			return util.MakeError(err, "OutboxGet")
		}
	}

	collection, err := actor.WantToServePage(page)
	if err != nil {
		return util.MakeError(err, "OutboxGet")
	}

	var offset = 15
	var pages []int
	pageLimit := (float64(collection.TotalItems) / float64(offset))

	if pageLimit > 11 {
		pageLimit = 11
	}

	for i := 0.0; i < pageLimit; i++ {
		pages = append(pages, int(i))
	}

	var data PageData
	data.Board.Name = actor.Name
	data.Board.PrefName = actor.PreferredUsername
	data.Board.Summary = actor.Summary
	data.Board.InReplyTo = ""
	data.Board.To = actor.Outbox
	data.Board.Actor = actor
	data.Board.ModCred, _ = util.GetPasswordFromSession(ctx)
	data.Board.Domain = config.Domain
	data.Board.Restricted = actor.Restricted
	data.CurrentPage = page
	data.ReturnTo = "feed"

	data.Board.Post.Actor = actor.Id

	capt, err := util.GetRandomCaptcha()
	if err != nil {
		return util.MakeError(err, "OutboxGet")
	}
	data.Board.Captcha = config.Domain + "/" + capt
	data.Board.CaptchaCode = post.GetCaptchaCode(data.Board.Captcha)

	data.Title = "/" + actor.Name + "/ - " + actor.PreferredUsername

	data.Key = config.Key

	data.Boards = webfinger.Boards
	data.Posts = collection.OrderedItems

	data.Pages = pages
	data.TotalPage = len(data.Pages) - 1

	data.Meta.Description = data.Board.Summary
	data.Meta.Url = data.Board.Actor.Id
	data.Meta.Title = data.Title

	data.Themes = &config.Themes
	data.ThemeCookie = getThemeCookie(ctx)

	return ctx.Render("nposts", fiber.Map{
		"page": data,
	}, "layouts/main")
}
