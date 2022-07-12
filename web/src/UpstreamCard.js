import * as React from 'react';
import Box from '@mui/material/Box';
import Card from '@mui/material/Card';
import CardActions from '@mui/material/CardActions';
import CardContent from '@mui/material/CardContent';
import Button from '@mui/material/Button';
import Typography from '@mui/material/Typography';

export default function UpstreamCard({
    // "address": "127.0.0.1:2828",
    // "healthy": false,
    // "num_requests": 1,
    // "fails": 0
    address,
    healthy,
    num_requests,
    fails,
}) {
  return (
    <Card sx={{ minWidth: 275 }}>
      <CardContent>
        <Typography sx={{ fontSize: 14 }} color="text.secondary" gutterBottom>
          Healthy: {healthy ? '👍' : '👎'}
        </Typography>
        <Typography variant="h5" component="div">
          Address: {address}
        </Typography>
        <Typography variant="body2">
          {`Requests: ${num_requests}`}
          <br />
          {`Fails: ${fails}`}
        </Typography>
      </CardContent>
    </Card>
  );
}