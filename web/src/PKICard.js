import * as React from 'react';
import Box from '@mui/material/Box';
import Card from '@mui/material/Card';
import CardActions from '@mui/material/CardActions';
import CardContent from '@mui/material/CardContent';
import Button from '@mui/material/Button';
import Typography from '@mui/material/Typography';
import Modal from '@mui/material/Modal';

export default function PKICard({
  // "id": "local",
  // "name": "Caddy Local Authority",
  // "root_common_name": "Caddy Local Authority - 2022 ECC Root",
  // "intermediate_common_name": "Caddy Local Authority - ECC Intermediate",
  // "root_certificate": "-----BEGIN CERTIFICATE-----\nMIIBozCCAUmgAwIBAgIQNIj3kiLEh1BxsfrnFNHhGTAKBggqhkjOPQQDAjAwMS4w\nLAYDVQQDEyVDYWRkeSBMb2NhbCBBdXRob3JpdHkgLSAyMDIyIEVDQyBSb290MB4X\nDTIyMDcwNjE2MDg1MFoXDTMyMDUxNDE2MDg1MFowMDEuMCwGA1UEAxMlQ2FkZHkg\nTG9jYWwgQXV0aG9yaXR5IC0gMjAyMiBFQ0MgUm9vdDBZMBMGByqGSM49AgEGCCqG\nSM49AwEHA0IABGskDyo6zB1VSd4vJfF3Zi2ds8FN4neL9SRAiG398zaOq4zCdrUK\nO61mW2ov+xmGI5yf9p6Y4NyOVBXoQc9PFsejRTBDMA4GA1UdDwEB/wQEAwIBBjAS\nBgNVHRMBAf8ECDAGAQH/AgEBMB0GA1UdDgQWBBQ/YOgPWyJ9gIs8zZH+Vc0D1Cfw\nGjAKBggqhkjOPQQDAgNIADBFAiEAqLylZDTag4JrDURDKLy2tn0UbGf4HqZhebTj\nb6bxOOYCIEsF8BbQnWoo4esZKjrjNw/RBF64whSRzrn6PQfeMDx2\n-----END CERTIFICATE-----\n",
  // "intermediate_certificate": "-----BEGIN CERTIFICATE-----\nMIIBxzCCAW2gAwIBAgIQYeM5swL4+nYPHCviawfspDAKBggqhkjOPQQDAjAwMS4w\nLAYDVQQDEyVDYWRkeSBMb2NhbCBBdXRob3JpdHkgLSAyMDIyIEVDQyBSb290MB4X\nDTIyMDcxMjA2MzM1NVoXDTIyMDcxOTA2MzM1NVowMzExMC8GA1UEAxMoQ2FkZHkg\nTG9jYWwgQXV0aG9yaXR5IC0gRUNDIEludGVybWVkaWF0ZTBZMBMGByqGSM49AgEG\nCCqGSM49AwEHA0IABOVnZQzlc6LHTzz+wcgn+m7j/9vhhy82uY3yAQ92U1RQnSnV\nbmVjAhNQaiC6ug4i8PSBXzTqhyk3bjC6qcroVvGjZjBkMA4GA1UdDwEB/wQEAwIB\nBjASBgNVHRMBAf8ECDAGAQH/AgEAMB0GA1UdDgQWBBTBgKIJ8wTeHnBjWN6yLb9n\nDxgWHzAfBgNVHSMEGDAWgBQ/YOgPWyJ9gIs8zZH+Vc0D1CfwGjAKBggqhkjOPQQD\nAgNIADBFAiEAvvO6jCI6ov60sT4wLUsh+2bp8MwVTUOx4vzy+NUg6b0CIFIXPqtO\n/K3QHNsjSjCc8JSL8cIx5H9dINLC/EwiJHzQ\n-----END CERTIFICATE-----\n"
  id,
  name,
  root_common_name,
  intermediate_common_name,
  root_certificate,
  intermediate_certificate,
}) {
  const [open, setOpen] = React.useState(false);
  const handleOpen = (name) => () => setOpen(name);
  const handleClose = () => setOpen(false);

  return (
    <Card sx={{ minWidth: 275 }}>
      <CardContent>
        <Typography sx={{ fontSize: 14 }} color="text.secondary" gutterBottom>
          ID: {id}
        </Typography>
        <Typography variant="h5" component="div">
          Name: {name}
        </Typography>
        <Typography variant="body2">
          {`Root CommonName:`}
          <br />
          {root_common_name}
        </Typography>
        <Typography variant="body2">
          {`Intermediate CommonName:`}
          <br />
          {intermediate_common_name}
        </Typography>
      </CardContent>
      <CardActions>
        <Button size="small" onClick={handleOpen('root')}>
          Show Root
        </Button>
        <Button size="small" onClick={handleOpen('intermediate')}>
          Show Intermediate
        </Button>
      </CardActions>
      <Modal
        open={open !== false}
        onClose={handleClose}
        aria-labelledby="modal-modal-title"
        aria-describedby="modal-modal-description"
      >
        <Box
          sx={{
            position: 'absolute',
            top: '50%',
            left: '50%',
            transform: 'translate(-50%, -50%)',
            width: 720,
            bgcolor: 'background.paper',
            border: '2px solid #000',
            boxShadow: 24,
            p: 4,
          }}
        >
          <Typography id="modal-modal-title" variant="h6" component="h2">
            {open}
          </Typography>
          <Typography id="modal-modal-description" sx={{ mt: 2 }} component="pre">
            {open === 'root' ? root_certificate : intermediate_certificate}
          </Typography>
        </Box>
      </Modal>
    </Card>
  );
}
