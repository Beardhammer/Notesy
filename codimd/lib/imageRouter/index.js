'use strict'

const Router = require('express').Router

const config = require('../config')
const logger = require('../logger')
const response = require('../response')

const imageRouter = module.exports = Router()

// upload image
imageRouter.post('/uploadimage', function (req, res) {
  var form = new (require('formidable').IncomingForm)()

  form.keepExtensions = true

  // Let formidable use system temp dir (don't set form.uploadDir)
  // We'll move the file to the correct location in the upload provider

  form.parse(req, function (err, fields, files) {
    if (err || !files.image || !files.image.path) {
      response.errorForbidden(req, res)
    } else {
      if (config.debug) {
        logger.info('SERVER received uploadimage: ' + JSON.stringify(files.image))
      }

      var noteAlias = (fields.noteAlias && typeof fields.noteAlias === 'string') ? fields.noteAlias : ''

      const uploadProvider = require('./' + config.imageUploadType)
      uploadProvider.uploadImage(files.image.path, noteAlias, function (err, url) {
        if (err !== null) {
          logger.error(err)
          return res.status(500).end('upload image error')
        }
        res.send({
          link: url
        })
      })
    }
  })
})
