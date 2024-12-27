package templates

const EmailTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>PDM Notes Registration Code</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            line-height: 1.6;
            color: #333333;
            margin: 0;
            padding: 0;
        }
        .container {
            max-width: 600px;
            margin: 0 auto;
            padding: 20px;
        }
        .header {
            text-align: center;
            padding: 20px 0;
            border-bottom: 2px solid #f0f0f0;
        }
        .content {
            padding: 30px 0;
        }
        .verification-code {
            background-color: #f7f7f7;
            border: 1px solid #e0e0e0;
            border-radius: 6px;
            padding: 20px;
            margin: 25px 0;
            text-align: center;
            font-size: 32px;
            letter-spacing: 5px;
            font-family: monospace;
        }
        .warning {
            color: #666666;
            font-size: 14px;
            margin: 20px 0;
        }
        .footer {
            text-align: center;
            padding: 20px 0;
            color: #666666;
            font-size: 12px;
            border-top: 1px solid #f0f0f0;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>PDM Notes Registration Code</h1>
        </div>
        <div class="content">
            <p>Hello,</p>
            <p>A verification code has been requested for your PDM Notes account registration. Please use the following code to verify your email address:</p>
            
            <div class="verification-code">
                {{.Code}}
            </div>
            
            <div class="warning">
                <p>If you didn't request this code, please ignore this email. For security reasons, this code will expire in 10 minutes.</p>
                <p>Please do not share this code with anyone. The PDM Notes team will never ask you for this code.</p>
            </div>
        </div>
        <div class="footer">
            <p>PDM Notes</p>
            <p>by Yi Yang</p>
            <p>&copy; 2024 PDM Notes. All rights reserved.</p>
        </div>
    </div>
</body>
</html>`
