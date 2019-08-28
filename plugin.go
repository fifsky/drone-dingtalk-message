package main

import (
	"errors"
	"fmt"
	"log"
	"strings"

	webhook "github.com/lddsb/dingtalk-webhook"
)

type (
	// Repo `repo base info`
	Repo struct {
		FullName string //  repository full name
	}

	// Build `build info`
	Build struct {
		Status string //  providers the current build status
		Link   string //  providers the current build link
	}

	// Commit `commit info`
	Commit struct {
		Branch  string //  providers the branch for the current commit
		Link    string //  providers the http link to the current commit in the remote source code management system(e.g.GitHub)
		Message string //  providers the commit message for the current build
		Sha     string //  providers the commit sha for the current build
		Authors CommitAuthors
	}

	// CommitAuthors `commit author info`
	CommitAuthors struct {
		Avatar string //  providers the author avatar for the current commit
		Email  string //  providers the author email for the current commit
		Name   string //  providers the author name for the current commit
	}

	// Drone `drone info`
	Drone struct {
		Repo   Repo
		Build  Build
		Commit Commit
	}

	// Config `plugin private config`
	Config struct {
		Debug       bool
		AccessToken string
		IsAtALL     bool
		Mobiles     string
		Username    string
		MsgType     string
	}

	// MessageConfig `DingTalk message struct`
	MessageConfig struct {
		ActionCard ActionCard
	}

	// ActionCard `action card message struct`
	ActionCard struct {
		LinkUrls       string
		LinkTitles     string
		HideAvatar     bool
		BtnOrientation bool
	}

	// Extra `extra variables`
	Extra struct {
		Color   ExtraColor
		Pic     ExtraPic
		LinkSha bool
	}

	// ExtraPic `extra config for pic`
	ExtraPic struct {
		WithPic       bool
		SuccessPicURL string
		FailurePicURL string
	}

	// ExtraColor `extra config for color`
	ExtraColor struct {
		WithColor    bool
		SuccessColor string
		FailureColor string
	}

	// Plugin `plugin all config`
	Plugin struct {
		Drone  Drone
		Config Config
		Extra  Extra
	}
)

// Exec `execute webhook`
func (p *Plugin) Exec() error {
	var err error
	if 0 == len(p.Config.AccessToken) {
		msg := "missing dingtalk access token"
		return errors.New(msg)
	}

	if 6 > len(p.Drone.Commit.Sha) {
		return errors.New("commit sha cannot short than 6")
	}

	newWebhook := webhook.NewWebHook(p.Config.AccessToken)
	mobiles := strings.Split(p.Config.Mobiles, ",")
	switch strings.ToLower(p.Config.MsgType) {
	case "markdown":
		err = newWebhook.SendMarkdownMsg("You have a new message...", p.baseTpl(), p.Config.IsAtALL, mobiles...)
	case "text":
		err = newWebhook.SendTextMsg(p.baseTpl(), p.Config.IsAtALL, mobiles...)
	case "link":
		err = newWebhook.SendLinkMsg(p.Drone.Build.Status, p.baseTpl(), p.Drone.Commit.Authors.Avatar, p.Drone.Build.Link)
	default:
		msg := "not support message type"
		err = errors.New(msg)
	}

	if err == nil {
		log.Println("send message success!")
	}

	return err
}

// markdownTpl `output the tpl of markdown`
func (p *Plugin) markdownTpl() string {
	var tpl string

	//  title
	title := fmt.Sprintf(" %s *Build %s* %s",
		strings.Title(p.Drone.Repo.FullName),
		strings.Title(p.Drone.Build.Status), p.getEmoticon())
	//  with color on title
	if p.Extra.Color.WithColor {
		title = fmt.Sprintf("<font color=%s>%s</font>", p.getColor(), title)
	}

	tpl = fmt.Sprintf("## %s \n", title)
	// with pic
	if p.Extra.Pic.WithPic {
		tpl += fmt.Sprintf("![%s](%s)\n\n",
			p.Drone.Build.Status,
			p.getPicURL())
	}

	//  commit message
	commitSha := p.Drone.Commit.Sha
	commitMsg := fmt.Sprintf("> %s %s %s", p.Drone.Commit.Branch, commitSha[:6], p.Drone.Commit.Message)
	if p.Extra.LinkSha {
		commitMsg = fmt.Sprintf("[%s](%s)", commitMsg, p.Drone.Commit.Link)
	}

	tpl += commitMsg + "\n"

	//  author info
	authorInfo := fmt.Sprintf("> Committer: %s`", p.Drone.Commit.Authors.Name)
	tpl += authorInfo + "\n"

	//  build detail link
	buildDetail := fmt.Sprintf("> [查看详情](%s)",
		p.Drone.Build.Link)
	tpl += buildDetail
	return tpl
}

func (p *Plugin) baseTpl() string {
	tpl := ""
	switch strings.ToLower(p.Config.MsgType) {
	case "markdown":
		tpl = p.markdownTpl()
	case "text":
		tpl = fmt.Sprintf(`[%s] %s
%s (%s)
@%s
%s (%s)
`,
			p.Drone.Build.Status,
			strings.TrimSpace(p.Drone.Commit.Message),
			p.Drone.Repo.FullName,
			p.Drone.Commit.Branch,
			p.Drone.Commit.Sha,
			p.Drone.Commit.Authors.Name,
			p.Drone.Commit.Authors.Email)
	case "link":
		tpl = fmt.Sprintf(`%s(%s) @%s %s(%s)`,
			p.Drone.Repo.FullName,
			p.Drone.Commit.Branch,
			p.Drone.Commit.Sha[:6],
			p.Drone.Commit.Authors.Name,
			p.Drone.Commit.Authors.Email)
	case "actionCard":
		//  coming soon

	}

	return tpl
}

/**
get emoticon
*/
func (p *Plugin) getEmoticon() string {
	emoticons := make(map[string]string)
	emoticons["success"] = "😀"
	emoticons["failure"] = "☹"

	emoticon, ok := emoticons[p.Drone.Build.Status]
	if ok {
		return emoticon
	}

	return "☹"
}

/**
get picture url
*/
func (p *Plugin) getPicURL() string {
	pics := make(map[string]string)
	//  success picture url
	pics["success"] = "https://static-cdn.verystar.net/s/imgs/success.png?x-oss-process=image/resize,h_100"
	if p.Extra.Pic.SuccessPicURL != "" {
		pics["success"] = p.Extra.Pic.SuccessPicURL
	}
	//  failure picture url
	pics["failure"] = "https://static-cdn.verystar.net/s/imgs/fail.png?x-oss-process=image/resize,h_100"
	if p.Extra.Pic.FailurePicURL != "" {
		pics["failure"] = p.Extra.Pic.FailurePicURL
	}

	url, ok := pics[p.Drone.Build.Status]
	if ok {
		return url
	}

	return ""
}

/**
get color for message title
*/
func (p *Plugin) getColor() string {
	colors := make(map[string]string)
	//  success color
	colors["success"] = "#008000"
	if p.Extra.Color.SuccessColor != "" {
		colors["success"] = "#" + p.Extra.Color.SuccessColor
	}
	//  failure color
	colors["failure"] = "#FF0000"
	if p.Extra.Color.FailureColor != "" {
		colors["failure"] = "#" + p.Extra.Color.FailureColor
	}

	color, ok := colors[p.Drone.Build.Status]
	if ok {
		return color
	}

	return ""
}
