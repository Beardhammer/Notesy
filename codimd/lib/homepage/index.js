'use strict'

const fs = require('fs')
const path = require('path')
const config = require('../config')
const { User } = require('../models')
const logger = require('../logger')
var moment = require('moment')

exports.showIndex = async (req, res) => {
  const isLogin = req.isAuthenticated()
  const deleteToken = ''

  const data = {
    signin: isLogin,
    infoMessage: req.flash('info'),
    errorMessage: req.flash('error'),
    privacyStatement: fs.existsSync(path.join(config.docsPath, 'privacy.md')),
    termsOfUse: fs.existsSync(path.join(config.docsPath, 'terms-of-use.md')),
    deleteToken: deleteToken
  }

  if (!isLogin) {
    const currentOpnote = moment(new Date()).format("YYYY-MM-DD")
    let serverURL = config.serverURL
    return res.redirect(301, serverURL + '/opnotes%2F' + currentOpnote + '?both')
  }

  const user = await User.findOne({
    where: {
      id: req.user.id
    }
  })
  if (user) {
    data.deleteToken = user.deleteToken
    return res.render('index.ejs', data)
  }

  logger.error(`error: user not found with id ${req.user.id}`)
  return res.render('index.ejs', data)
}
