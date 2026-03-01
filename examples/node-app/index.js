const express = require('express');
const { Pool } = require('pg');

const app = express();
const PORT = 3000;

// PostgreSQL connection pool
const pool = new Pool({
  host: 'localhost',
  port: 5432,
  user: 'dev',
  password: 'devpass',
  database: 'devdb'
});

// Test database connection on startup
pool.query('SELECT NOW()', (err, res) => {
  if (err) {
    console.error(' PostgreSQL connection failed:', err.message);
  } else {
    console.log(' PostgreSQL connected at:', res.rows[0].now);
  }
});

// Routes
app.get('/', (req, res) => {
  res.send(`
    <!DOCTYPE html>
    <html>
    <head>
      <title>DependencyResolver - Node.js Demo</title>
      <style>
        body {
          font-family: Arial, sans-serif;
          max-width: 800px;
          margin: 50px auto;
          padding: 20px;
          background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
          color: white;
        }
        .container {
          background: rgba(255, 255, 255, 0.1);
          padding: 30px;
          border-radius: 10px;
          backdrop-filter: blur(10px);
        }
        h1 { margin-top: 0; }
        .check { color: #4ade80; font-size: 1.2em; }
        a { color: #fbbf24; text-decoration: none; }
        a:hover { text-decoration: underline; }
        code { 
          background: rgba(0,0,0,0.3); 
          padding: 2px 6px; 
          border-radius: 3px;
        }
        .endpoint {
          background: rgba(255,255,255,0.1);
          padding: 10px;
          margin: 10px 0;
          border-radius: 5px;
        }
      </style>
    </head>
    <body>
      <div class="container">
        <h1>🚀 DependencyResolver - Node.js Demo</h1>
        
        <p class="check"> Express.js is working!</p>
        <p class="check"> PostgreSQL driver (pg) loaded!</p>
        <p class="check"> Node.js ${process.version}</p>
        
        <hr style="border: 1px solid rgba(255,255,255,0.3); margin: 20px 0;">
        
        <h2> API Endpoints:</h2>
        
        <div class="endpoint">
          <strong><a href="/api/status">/api/status</a></strong>
          <p>Check database connection status</p>
        </div>
        
        <div class="endpoint">
          <strong><a href="/api/users">/api/users</a></strong>
          <p>Create sample table and fetch users</p>
        </div>
        
        <div class="endpoint">
          <strong><a href="/api/info">/api/info</a></strong>
          <p>Show system information</p>
        </div>
        
        <hr style="border: 1px solid rgba(255,255,255,0.3); margin: 20px 0;">
        
        <p><small> Auto-detected and bootstrapped by DependencyResolver</small></p>
      </div>
    </body>
    </html>
  `);
});

// API: Check database status
app.get('/api/status', async (req, res) => {
  try {
    const result = await pool.query('SELECT NOW() as time, version() as version');
    res.json({
      status: 'connected',
      timestamp: result.rows[0].time,
      database: result.rows[0].version
    });
  } catch (error) {
    res.status(500).json({
      status: 'error',
      message: error.message
    });
  }
});

// API: Create table and fetch users
app.get('/api/users', async (req, res) => {
  try {
    // Create table if not exists
    await pool.query(`
      CREATE TABLE IF NOT EXISTS users (
        id SERIAL PRIMARY KEY,
        name VARCHAR(100),
        email VARCHAR(100),
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
      )
    `);

    
    const countResult = await pool.query('SELECT COUNT(*) FROM users');
    if (parseInt(countResult.rows[0].count) === 0) {
      await pool.query(`
        INSERT INTO users (name, email) VALUES
        ('Alice Johnson', 'alice@example.com'),
        ('Bob Smith', 'bob@example.com'),
        ('Charlie Brown', 'charlie@example.com')
      `);
    }

    // Fetch all users
    const result = await pool.query('SELECT * FROM users ORDER BY id');
    
    res.json({
      success: true,
      count: result.rows.length,
      users: result.rows
    });
  } catch (error) {
    res.status(500).json({
      success: false,
      error: error.message
    });
  }
});

// API: System info
app.get('/api/info', (req, res) => {
  res.json({
    node_version: process.version,
    platform: process.platform,
    arch: process.arch,
    memory: {
      total: Math.round(require('os').totalmem() / 1024 / 1024) + ' MB',
      free: Math.round(require('os').freemem() / 1024 / 1024) + ' MB'
    },
    uptime: Math.round(process.uptime()) + ' seconds'
  });
});

// 404 handler
app.use((req, res) => {
  res.status(404).send(`
    <h1>404 - Not Found</h1>
    <p>Go back to <a href="/">home page</a></p>
  `);
});

// Start server
app.listen(PORT, () => {
  console.log('━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━');
  console.log(` Server running on http://localhost:${PORT}`);
  console.log('━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━');
  console.log('');
  console.log(' Available endpoints:');
  console.log(`   → http://localhost:${PORT}/`);
  console.log(`   → http://localhost:${PORT}/api/status`);
  console.log(`   → http://localhost:${PORT}/api/users`);
  console.log(`   → http://localhost:${PORT}/api/info`);
  console.log('');
  console.log('Press Ctrl+C to stop');
  console.log('━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━');
});

// Graceful shutdown
process.on('SIGTERM', () => {
  console.log('\n Shutting down gracefully...');
  pool.end(() => {
    console.log(' Database pool closed');
    process.exit(0);
  });
});