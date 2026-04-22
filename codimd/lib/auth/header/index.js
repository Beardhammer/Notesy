'use strict'

const Router = require('express').Router
const passport = require('passport')
const { Strategy: CustomStrategy } = require('passport-custom')

const config = require('../../config')
const models = require('../../models')
const { setReturnToFromReferer } = require('../utils')

const authRouter = module.exports = Router()

passport.use('header', new CustomStrategy(async (req, done) => {
  try {
    const nameHdr = (config.header.usernameHeader || 'X-Authentik-Username').toLowerCase()
    const emailHdr = (config.header.emailHeader || 'X-Authentik-Email').toLowerCase()
    const username = req.headers[nameHdr]
    if (!username) return done(null, false, { message: 'no auth header' })

    const email = req.headers[emailHdr] || ''
    const profile = {
      id: `header:${username}`,
      username,
      displayName: username,
      emails: email ? [{ value: email }] : []
    }

    const stringifiedProfile = JSON.stringify(profile)
    const [user] = await models.User.findOrCreate({
      where: { profileid: profile.id },
      defaults: { profile: stringifiedProfile }
    })
    // Refresh the stored profile if email or other attrs change between logins.
    if (user.profile !== stringifiedProfile) {
      user.profile = stringifiedProfile
      await user.save()
    }
    return done(null, user)
  } catch (err) { return done(err) }
}))

authRouter.use((req, res, next) => {
  if (req.isAuthenticated()) return next()
  setReturnToFromReferer(req)
  passport.authenticate('header', { session: true })(req, res, next)
})
