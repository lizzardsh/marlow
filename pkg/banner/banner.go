// Banner
package banner

import "marlow/pkg/terminal"

const Art = `
• ▌ ▄ ·.  ▄▄▄· ▄▄▄  ▄▄▌        ▄▄▌ ▐ ▄▌
·██ ▐███▪▐█ ▀█ ▀▄ █·██•  ▪     ██· █▌▐█
▐█ ▌▐▌▐█·▄█▀▀█ ▐▀▀▄ ██▪   ▄█▀▄ ██▪▐█▐▐▌
██ ██▌▐█▌▐█ ▪▐▌▐█•█▌▐█▌▐▌▐█▌.▐▌▐█▌██▐█▌
▀▀  █▪▀▀▀ ▀  ▀ .▀  ▀.▀▀▀  ▀█▄▀▪ ▀▀▀▀ ▀▪
`

func Render() string {
	return terminal.Gradient(terminal.BannerGradientStart, terminal.BannerGradientEnd, Art)
}
