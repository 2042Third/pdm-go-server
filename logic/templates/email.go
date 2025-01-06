package templates

const EmailTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta name="color-scheme" content="light dark">
    <meta name="supported-color-schemes" content="light dark">
    <title>PDM Notes Email Verification Code</title>
    <!--[if mso]>
    <noscript>
        <xml>
            <o:OfficeDocumentSettings>
                <o:PixelsPerInch>96</o:PixelsPerInch>
            </o:OfficeDocumentSettings>
        </xml>
    </noscript>
    <![endif]-->
    <style>
        /* Reset styles for email clients */
        body, table, td, p, a {
            -webkit-text-size-adjust: 100%;
            -ms-text-size-adjust: 100%;
            margin: 0;
            padding: 0;
        }

        /* Base styles */
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif;
            line-height: 1.6;
            color: #333333;
            background-color: #f9f9f9;
        }

        .container {
            max-width: 600px;
            margin: 0 auto;
            padding: 20px;
            background-color: #ffffff;
            border-radius: 8px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }

        .header {
            text-align: center;
            padding: 20px 0;
            border-bottom: 2px solid #f0f0f0;
        }

        .header h1 {
            color: #2563eb;
            font-size: 24px;
            margin: 0;
            padding: 0;
        }

        .content {
            padding: 30px 20px;
            background-color: #ffffff;
        }

        .verification-code {
            background-color: #f3f4f6;
            border: 1px solid #e5e7eb;
            border-radius: 8px;
            padding: 20px;
            margin: 25px 0;
            text-align: center;
            font-size: 32px;
            letter-spacing: 8px;
            font-family: 'Courier New', monospace;
            font-weight: bold;
            color: #1f2937;
        }

        .warning {
            color: #6b7280;
            font-size: 14px;
            margin: 20px 0;
            padding: 15px;
            background-color: #fff9f9;
            border-left: 4px solid #ef4444;
            border-radius: 4px;
        }

        .footer {
            text-align: center;
            padding: 20px 0;
            color: #6b7280;
            font-size: 12px;
            border-top: 1px solid #f0f0f0;
            background-color: #f9fafb;
        }

        /* Dark mode support */
        @media (prefers-color-scheme: dark) {
            body {
                background-color: #1f2937;
                color: #f9fafb;
            }
            .container {
                background-color: #374151;
            }
            .verification-code {
                background-color: #4b5563;
                border-color: #6b7280;
                color: #f9fafb;
            }
            .warning {
                background-color: #422424;
                border-left-color: #dc2626;
                color: #fca5a5;
            }
            .footer {
                background-color: #374151;
                color: #9ca3af;
            }
        }

        /* Mobile responsiveness */
        @media screen and (max-width: 600px) {
            .container {
                width: 100% !important;
                margin: 0 !important;
            }
            .content {
                padding: 20px 15px !important;
            }
            .verification-code {
                font-size: 24px !important;
                letter-spacing: 6px !important;
            }
        }
    </style>
</head>
<body>
    <table role="presentation" width="100%" cellpadding="0" cellspacing="0" border="0">
        <tr>
            <td align="center" style="padding: 20px 0;">
                <div class="container">
                    <div class="header">
                        <h1>PDM Notes Verification</h1>
                    </div>
                    <div class="content">
                        <p>Hello,</p>
                        <p>Your verification code is:</p>
                        
                        <div class="verification-code">
                            {{.Code}}
                        </div>
                        
                        <div class="warning">
                            <p>This code will expire in 10 minutes.</p>
                            <p>If you didn't request this code, please ignore this email.</p>
                            <p>PDM Notes will never ask you for this code.</p>
                        </div>
                    </div>
                    <div class="footer">
                        <p>PDM Notes - Secure Note-Taking Platform</p>
                        <p>© 2024 PDM Notes. All rights reserved.</p>
                        <p style="margin-top: 10px; font-size: 11px;">
                            This is an automated message, please do not reply.
                        </p>
                    </div>
                </div>
            </td>
        </tr>
    </table>
</body>
</html>`
