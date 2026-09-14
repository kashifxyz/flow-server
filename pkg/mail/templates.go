package mail

import "fmt"

const (
	TemplateVerifyEmail     = "verify_email"
	TemplateResetPassword   = "reset_password"
	TemplateWorkspaceInvite = "workspace_invite"
)

func subject(msg Message) string {
	switch msg.Template {
	case TemplateVerifyEmail:
		return "Verify your Flow email"
	case TemplateResetPassword:
		return "Reset your Flow password"
	case TemplateWorkspaceInvite:
		return fmt.Sprintf("You've been invited to %s on Flow", payloadString(msg, "workspace"))
	case "":
		return "Flow"
	default:
		return "Flow"
	}
}

func render(msg Message) string {
	link := payloadString(msg, "link")
	expires := payloadString(msg, "expires")
	switch msg.Template {
	case TemplateVerifyEmail:
		return fmt.Sprintf("Verify your Flow account by opening this link:\n\n%s\n\nThis link expires in %s.\nIf you did not create an account, ignore this email.\n", link, expires)
	case TemplateResetPassword:
		return fmt.Sprintf("Reset your Flow password by opening this link:\n\n%s\n\nThis link expires in %s.\nIf you did not request a reset, ignore this email.\n", link, expires)
	case TemplateWorkspaceInvite:
		workspace := payloadString(msg, "workspace")
		return fmt.Sprintf("You've been invited to join \"%s\" on Flow.\n\nAccept the invite by opening this link:\n\n%s\n\nIf you weren't expecting this, you can ignore this email.\n", workspace, link)
	default:
		if link != "" {
			return link
		}
		return "Flow"
	}
}

func payloadString(msg Message, key string) string {
	if msg.Payload == nil {
		return ""
	}
	s, _ := msg.Payload[key].(string)
	return s
}
