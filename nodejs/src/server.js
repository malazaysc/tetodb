/**
 * MiniLiteDB Express Demo Server
 *
 * This demonstrates how to use the MiniLiteDB WASM database
 * in a Node.js Express application.
 */

const express = require('express');
const { MiniLiteDB } = require('./minilite');

const app = express();
const PORT = process.env.PORT || 3000;
const DB_PATH = './demo.db';

// Middleware
app.use(express.json());
app.use(express.urlencoded({ extended: true }));

// Initialize database
let db;
let usersCollection;

async function initDatabase() {
  console.log('Initializing MiniLiteDB...');
  db = new MiniLiteDB();
  await db.open(DB_PATH);
  usersCollection = db.collection('users');
  console.log('Database initialized successfully!');
}

// Logging middleware
app.use((req, res, next) => {
  console.log(`${new Date().toISOString()} - ${req.method} ${req.path}`);
  next();
});

// Routes

/**
 * GET / - Welcome message
 */
app.get('/', (req, res) => {
  res.json({
    message: 'Welcome to MiniLiteDB Demo API',
    endpoints: {
      'GET /': 'This message',
      'GET /users': 'List all users',
      'GET /users/:id': 'Get a user by ID',
      'POST /users': 'Create a new user',
      'PATCH /users/:id': 'Update a user',
      'DELETE /users/:id': 'Delete a user',
      'GET /stats': 'Get database statistics',
      'POST /compact': 'Compact the database',
    },
  });
});

/**
 * GET /users - List all users
 * Query params:
 *   - name: Filter by name
 *   - email: Filter by email
 */
app.get('/users', async (req, res) => {
  try {
    const filter = {};

    // Apply filters if provided
    if (req.query.name) {
      filter.name = req.query.name;
    }
    if (req.query.email) {
      filter.email = req.query.email;
    }

    const users = await usersCollection.find(filter);

    res.json({
      success: true,
      count: users.length,
      users,
    });
  } catch (error) {
    console.error('Error listing users:', error);
    res.status(500).json({
      success: false,
      error: error.message,
    });
  }
});

/**
 * GET /users/:id - Get a user by ID
 */
app.get('/users/:id', async (req, res) => {
  try {
    const user = await usersCollection.findById(req.params.id);

    if (!user) {
      return res.status(404).json({
        success: false,
        error: 'User not found',
      });
    }

    res.json({
      success: true,
      user,
    });
  } catch (error) {
    console.error('Error getting user:', error);
    res.status(500).json({
      success: false,
      error: error.message,
    });
  }
});

/**
 * POST /users - Create a new user
 * Body:
 *   - name: User's name (required)
 *   - email: User's email (required)
 *   - age: User's age (optional)
 *   - role: User's role (optional)
 */
app.post('/users', async (req, res) => {
  try {
    const { name, email, age, role } = req.body;

    // Validation
    if (!name || !email) {
      return res.status(400).json({
        success: false,
        error: 'Name and email are required',
      });
    }

    // Create user document
    const user = {
      name,
      email,
      age: age || null,
      role: role || 'user',
      createdAt: new Date().toISOString(),
    };

    // Insert into database
    const id = await usersCollection.insert(user);

    res.status(201).json({
      success: true,
      message: 'User created successfully',
      id,
      user: { ...user, id },
    });
  } catch (error) {
    console.error('Error creating user:', error);
    res.status(500).json({
      success: false,
      error: error.message,
    });
  }
});

/**
 * PATCH /users/:id - Update a user
 * Body: Fields to update (name, email, age, role, etc.)
 */
app.patch('/users/:id', async (req, res) => {
  try {
    const { id } = req.params;

    // Check if user exists
    const existingUser = await usersCollection.findById(id);
    if (!existingUser) {
      return res.status(404).json({
        success: false,
        error: 'User not found',
      });
    }

    // Prepare update object
    const update = { ...req.body };
    update.updatedAt = new Date().toISOString();

    // Don't allow updating the ID
    delete update.id;

    // Update user
    await usersCollection.updateById(id, update);

    // Fetch updated user
    const updatedUser = await usersCollection.findById(id);

    res.json({
      success: true,
      message: 'User updated successfully',
      user: updatedUser,
    });
  } catch (error) {
    console.error('Error updating user:', error);
    res.status(500).json({
      success: false,
      error: error.message,
    });
  }
});

/**
 * DELETE /users/:id - Delete a user
 */
app.delete('/users/:id', async (req, res) => {
  try {
    const { id } = req.params;

    // Check if user exists
    const existingUser = await usersCollection.findById(id);
    if (!existingUser) {
      return res.status(404).json({
        success: false,
        error: 'User not found',
      });
    }

    // Delete user
    await usersCollection.deleteById(id);

    res.json({
      success: true,
      message: 'User deleted successfully',
      user: existingUser,
    });
  } catch (error) {
    console.error('Error deleting user:', error);
    res.status(500).json({
      success: false,
      error: error.message,
    });
  }
});

/**
 * GET /stats - Get database statistics
 */
app.get('/stats', async (req, res) => {
  try {
    const stats = await db.stats();
    const userCount = await usersCollection.count();

    res.json({
      success: true,
      stats: {
        ...stats,
        userCount,
      },
    });
  } catch (error) {
    console.error('Error getting stats:', error);
    res.status(500).json({
      success: false,
      error: error.message,
    });
  }
});

/**
 * POST /compact - Compact the database
 */
app.post('/compact', async (req, res) => {
  try {
    await db.compact();

    res.json({
      success: true,
      message: 'Database compacted successfully',
    });
  } catch (error) {
    console.error('Error compacting database:', error);
    res.status(500).json({
      success: false,
      error: error.message,
    });
  }
});

// Error handling middleware
app.use((err, req, res, next) => {
  console.error('Unhandled error:', err);
  res.status(500).json({
    success: false,
    error: 'Internal server error',
  });
});

// 404 handler
app.use((req, res) => {
  res.status(404).json({
    success: false,
    error: 'Endpoint not found',
  });
});

// Graceful shutdown
process.on('SIGINT', async () => {
  console.log('\nShutting down gracefully...');

  if (db) {
    await db.close();
    console.log('Database closed');
  }

  process.exit(0);
});

process.on('SIGTERM', async () => {
  console.log('\nShutting down gracefully...');

  if (db) {
    await db.close();
    console.log('Database closed');
  }

  process.exit(0);
});

// Start server
async function startServer() {
  try {
    await initDatabase();

    app.listen(PORT, () => {
      console.log(`\n========================================`);
      console.log(`MiniLiteDB Demo Server`);
      console.log(`========================================`);
      console.log(`Server running on http://localhost:${PORT}`);
      console.log(`Database file: ${DB_PATH}`);
      console.log(`\nAvailable endpoints:`);
      console.log(`  GET    /users          - List all users`);
      console.log(`  GET    /users/:id      - Get user by ID`);
      console.log(`  POST   /users          - Create a user`);
      console.log(`  PATCH  /users/:id      - Update a user`);
      console.log(`  DELETE /users/:id      - Delete a user`);
      console.log(`  GET    /stats          - Database statistics`);
      console.log(`  POST   /compact        - Compact database`);
      console.log(`========================================\n`);
    });
  } catch (error) {
    console.error('Failed to start server:', error);
    process.exit(1);
  }
}

startServer();
