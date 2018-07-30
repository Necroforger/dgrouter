# Exmiddleware
Collection of middleware for common things

## Example
```go
router.Group(func(r *exrouter.Route) {
	// Default callback function to use when a middleware has an error
	// It will reply to the sender with the error that occured
	reply := exmiddleware.CatchDefault

	// Create a specific reply for when a middleware errors
	replyNSFW := exmiddleware.CatchReply("You need to be in an NSFW channel to use this command")

	r.Use(
		// Inserts the calling member into ctx.Data
		exmiddleware.GetMember(reply),

		// Inserts the Guild in which the message came from into ctx.Data
		exmiddleware.GetGuild(reply),

		// Require that the message originates from an nsfw channel
		// If there is an error, run the function replyNSFW
		exmiddleware.RequireNSFW(replyNSFW),

		// Enforce that the commands in this group can only be used once per 10 seconds per user
		exmiddleware.UserCooldown(time.Second*10, exmiddleware.CatchReply("This command is on cooldown...")),
	)

	r.On("nsfw", func(ctx *exrouter.Context) {
		ctx.Reply("This command was used in a NSFW channel\n" +
			"Your guild is: " + exmiddleware.Guild(ctx).Name + "\n" +
			"Your nickanme is: " + exmiddleware.Member(ctx).Nick,
		)
	})
})
```