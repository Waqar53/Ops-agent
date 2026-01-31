const express = require('express');
const app = express();
app.use(express.json());
app.get('/health', (req, res) => res.json({ status: 'healthy' }));
app.get('/api/users', (req, res) => res.json([]));
app.listen(3000, () => console.log('Server running'));
