/**
 * MiniLiteDB - A tiny embeddable NoSQL database compiled to WebAssembly
 *
 * This module provides a JavaScript wrapper around the Go WASM implementation
 * of MiniLiteDB, offering a clean Promise-based API for Node.js applications.
 */

const fs = require('fs');
const path = require('path');

// Load the Go WASM runtime
require('../wasm/wasm_exec.js');

/**
 * MiniLiteDB class - Main database interface
 */
class MiniLiteDB {
  constructor() {
    this.isOpen = false;
    this.dbPath = null;
    this.wasmInstance = null;
  }

  /**
   * Initialize and load the WASM module
   * This must be called before opening a database
   */
  async init() {
    if (this.wasmInstance) {
      return; // Already initialized
    }

    const wasmPath = path.join(__dirname, '../wasm/minilite.wasm');
    const wasmBuffer = fs.readFileSync(wasmPath);

    const go = new Go();
    const result = await WebAssembly.instantiate(wasmBuffer, go.importObject);

    // Run the Go runtime
    go.run(result.instance);

    this.wasmInstance = result.instance;

    // Wait a bit for Go to register functions
    await new Promise(resolve => setTimeout(resolve, 100));
  }

  /**
   * Open a database at the specified path
   * Creates the database if it doesn't exist
   *
   * @param {string} dbPath - Path to the database file
   * @returns {Promise<MiniLiteDB>} - Returns this for chaining
   */
  async open(dbPath) {
    if (!this.wasmInstance) {
      await this.init();
    }

    const result = miniLiteOpen(dbPath);

    if (!result.success) {
      throw new Error(result.error);
    }

    this.isOpen = true;
    this.dbPath = dbPath;

    return this;
  }

  /**
   * Get a collection by name
   *
   * @param {string} name - Collection name
   * @returns {Collection} - Collection instance
   */
  collection(name) {
    if (!this.isOpen) {
      throw new Error('Database is not open');
    }

    return new Collection(name, this);
  }

  /**
   * Get database statistics
   *
   * @returns {Promise<object>} - Database stats
   */
  async stats() {
    this._checkOpen();

    const result = miniLiteStats();

    if (!result.success) {
      throw new Error(result.error);
    }

    return result.stats;
  }

  /**
   * Compact the database file
   * Removes deleted/updated records and reclaims disk space
   *
   * @returns {Promise<void>}
   */
  async compact() {
    this._checkOpen();

    const result = miniLiteCompact();

    if (!result.success) {
      throw new Error(result.error);
    }
  }

  /**
   * Close the database
   *
   * @returns {Promise<void>}
   */
  async close() {
    if (!this.isOpen) {
      return;
    }

    const result = miniLiteClose();

    if (!result.success) {
      throw new Error(result.error);
    }

    this.isOpen = false;
    this.dbPath = null;
  }

  /**
   * Internal helper to check if database is open
   * @private
   */
  _checkOpen() {
    if (!this.isOpen) {
      throw new Error('Database is not open');
    }
  }
}

/**
 * Collection class - Represents a collection of documents
 */
class Collection {
  constructor(name, db) {
    this.name = name;
    this.db = db;
  }

  /**
   * Insert a document into the collection
   *
   * @param {object} document - The document to insert
   * @returns {Promise<string>} - The inserted document's ID
   */
  async insert(document) {
    this.db._checkOpen();

    const jsonDoc = JSON.stringify(document);
    const result = miniLiteInsert(this.name, jsonDoc);

    if (!result.success) {
      throw new Error(result.error);
    }

    return result.id;
  }

  /**
   * Insert multiple documents
   *
   * @param {Array<object>} documents - Array of documents to insert
   * @returns {Promise<Array<string>>} - Array of inserted document IDs
   */
  async insertMany(documents) {
    const ids = [];

    for (const doc of documents) {
      const id = await this.insert(doc);
      ids.push(id);
    }

    return ids;
  }

  /**
   * Find documents matching a filter
   *
   * @param {object} filter - Filter criteria (optional)
   * @returns {Promise<Array<object>>} - Array of matching documents
   */
  async find(filter = {}) {
    this.db._checkOpen();

    const filterJSON = Object.keys(filter).length > 0 ? JSON.stringify(filter) : '';
    const result = miniLiteFind(this.name, filterJSON);

    if (!result.success) {
      throw new Error(result.error);
    }

    return JSON.parse(result.documents);
  }

  /**
   * Find a single document by ID
   *
   * @param {string} id - Document ID
   * @returns {Promise<object|null>} - The document or null if not found
   */
  async findById(id) {
    this.db._checkOpen();

    const result = miniLiteFindByID(this.name, id);

    if (!result.success) {
      // Document not found
      return null;
    }

    return JSON.parse(result.document);
  }

  /**
   * Find the first document matching a filter
   *
   * @param {object} filter - Filter criteria
   * @returns {Promise<object|null>} - The first matching document or null
   */
  async findOne(filter = {}) {
    const docs = await this.find(filter);
    return docs.length > 0 ? docs[0] : null;
  }

  /**
   * Update a document by ID
   *
   * @param {string} id - Document ID
   * @param {object} update - Fields to update
   * @returns {Promise<void>}
   */
  async updateById(id, update) {
    this.db._checkOpen();

    const updateJSON = JSON.stringify(update);
    const result = miniLiteUpdate(this.name, id, updateJSON);

    if (!result.success) {
      throw new Error(result.error);
    }
  }

  /**
   * Update the first document matching a filter
   *
   * @param {object} filter - Filter criteria
   * @param {object} update - Fields to update
   * @returns {Promise<boolean>} - True if a document was updated
   */
  async updateOne(filter, update) {
    const doc = await this.findOne(filter);

    if (!doc) {
      return false;
    }

    await this.updateById(doc.id, update);
    return true;
  }

  /**
   * Delete a document by ID
   *
   * @param {string} id - Document ID
   * @returns {Promise<void>}
   */
  async deleteById(id) {
    this.db._checkOpen();

    const result = miniLiteDelete(this.name, id);

    if (!result.success) {
      throw new Error(result.error);
    }
  }

  /**
   * Delete the first document matching a filter
   *
   * @param {object} filter - Filter criteria
   * @returns {Promise<boolean>} - True if a document was deleted
   */
  async deleteOne(filter) {
    const doc = await this.findOne(filter);

    if (!doc) {
      return false;
    }

    await this.deleteById(doc.id);
    return true;
  }

  /**
   * Delete all documents matching a filter
   *
   * @param {object} filter - Filter criteria
   * @returns {Promise<number>} - Number of documents deleted
   */
  async deleteMany(filter) {
    const docs = await this.find(filter);

    for (const doc of docs) {
      await this.deleteById(doc.id);
    }

    return docs.length;
  }

  /**
   * Count documents in the collection
   *
   * @param {object} filter - Filter criteria (optional)
   * @returns {Promise<number>} - Number of documents
   */
  async count(filter = {}) {
    this.db._checkOpen();

    const filterJSON = Object.keys(filter).length > 0 ? JSON.stringify(filter) : '';
    const result = miniLiteCount(this.name, filterJSON);

    if (!result.success) {
      throw new Error(result.error);
    }

    return result.count;
  }
}

module.exports = { MiniLiteDB, Collection };
