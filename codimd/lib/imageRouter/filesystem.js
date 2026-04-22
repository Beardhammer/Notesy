'use strict'
const fs = require('fs')
const path = require('path')

const config = require('../config')
const logger = require('../logger')

exports.uploadImage = function (imagePath, noteAlias, callback) {
  if (!imagePath || typeof imagePath !== 'string') {
    callback(new Error('Image path is missing or wrong'), null)
    return
  }

  if (!callback || typeof callback !== 'function') {
    logger.error('Callback has to be a function')
    return
  }

  // Determine target directory based on note alias
  var targetDir
  if (noteAlias) {
    var noteDir = path.dirname(path.join(config.docsPath, noteAlias))
    targetDir = path.join(noteDir, 'images')
  } else {
    targetDir = path.join(config.docsPath, 'images')
  }

  // Create directory if it doesn't exist
  if (!fs.existsSync(targetDir)) {
    try {
      fs.mkdirSync(targetDir, { recursive: true })
    } catch (e) {
      callback(new Error('Failed to create images directory: ' + e.message), null)
      return
    }
  }

  // Move file from temp to target directory
  var filename = path.basename(imagePath)
  var targetPath = path.join(targetDir, filename)

  try {
    // Use copy+delete instead of rename to handle cross-filesystem moves
    fs.copyFileSync(imagePath, targetPath)
    fs.unlinkSync(imagePath)
  } catch (e) {
    callback(new Error('Failed to move uploaded file: ' + e.message), null)
    return
  }

  // Build URL relative to docs path
  var relativePath = path.relative(config.docsPath, targetPath).split(path.sep).join('/')
  var url = config.serverURL + '/docs/' + relativePath
  callback(null, url)
}
