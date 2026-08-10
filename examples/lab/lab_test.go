package lab_test

import (
	"testing"

	"github.com/sidwiskers/hermes"
	"github.com/sidwiskers/hermes/testkit"
)

func TestSettingsConversation(t *testing.T) {
	lab := testkit.NewLab(t)

	lab.Bot.Command("settings", func(c *hermes.Context) error {
		keyboard := hermes.Keyboard(hermes.Row(hermes.Button("Save", "save")))
		return c.Send("Settings", hermes.WithKeyboard(keyboard))
	})
	lab.Bot.Callback("save", func(c *hermes.Context) error {
		if err := c.Acknowledge(); err != nil {
			return err
		}
		return c.Edit("Saved")
	})

	alice := lab.PrivateUser(42, "alice")
	alice.Command("settings").Want(testkit.Sent("Settings"))
	alice.Callback("save").Want(
		testkit.Acknowledged(),
		testkit.Edited("Saved"),
	)
}
