import fs from "fs";
import dgram from "node:dgram";
import path from "path";
import cron from "node-cron";

interface FileData {
    filename: string;
    filesize: number;
    currentbyte: number;
    data: string;
}``

const server = dgram.createSocket("udp4");

server.on("error", (err) => {
    console.log(`server error:\n${err.stack}`);
    server.close();
});

server.on("listening", () => {
    const address = server.address();
    console.log(`server listening ${address.address}:${address.port}`);
});

server.on("message", async (msg, rinfo) => {
    // console.log(`server got: ${msg} from ${rinfo.address}:${rinfo.port}`);
    const fileData = JSON.parse(msg.toString("utf8")) as FileData;

    console.log("Buffer: ", fileData);
    console.log("Buffer Size: ", msg.length);
    console.log("RInfo: ", rinfo);

    //creating file from the buffer
    if (fileData.filename && fileData.data && fileData.currentbyte <= fileData.filesize) {
        const filePath = path.join(__dirname, "..", "rcvd-files", fileData.filename);
        let fileStream: fs.WriteStream;

        if (fs.existsSync(filePath)) {
            fileStream = fs.createWriteStream(filePath, {
                flags: "a",
                highWaterMark: 1024 * 1024, //1MB   - to avoid memory issues with large files - it will drain the buffer after 1MB  
            });
        } else {
            fileStream = fs.createWriteStream(filePath, {
                flags: "w",
                highWaterMark: 1024 * 1024, //1MB   - to avoid memory issues with large files - it will drain the buffer after 1MB  
            });
        }

        //writing buffer to file in base64 format to avoid encoding issues with special characters in binary data like images or pdfs
        if (fileStream.write(Buffer.from(fileData.data, "base64"), (err) => {
            if (err) {
                console.log("Error writing to file: ", err);
            } else {
                // Closing file stream if all data is received
                if (fileData.currentbyte === fileData.filesize) {
                    fileStream.end();
                    console.log("File received successfully!");
                }


                //send the acknowledgement to the client    
                server.send(Buffer.from("ACK"), rinfo.port, rinfo.address, (err) => {
                    if (err) {
                        console.log(`server error:\n${err.stack}`);
                    }

                    console.log(`server sent: ACK`);
                })
            }
        })) {
            console.log("Data written successfully!");
        }

        //closing file stream if all data is received

        fileStream.on("error", (err) => {
            console.log("Error writing to file: ", err);
        });

        fileStream.on("close", () => {
            console.log("File stream closed successfully!");
        });

        fileStream.on("drain", () => {
            console.log("File stream drained successfully!");
        })
    }

    console.log("=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=");
});

server.bind(41234);

//register a cron job to delete files older than 30mins
cron.schedule("*/1 * * * *", (now) => { //every 5mins
    console.log("Running cron job to delete files older than 2mins: ", now);

    const dirPath = path.join(__dirname, "..", "rcvd-files");

    const readDir = fs.readdirSync(dirPath);

    readDir.forEach((file) => {
        const filePath = path.join(dirPath, file);
        const fileStat = fs.statSync(filePath);

        if (fileStat.isFile() && (Date.now() - fileStat.mtimeMs) > 2 * 60 * 1000) {
            fs.unlinkSync(filePath);
            console.log("File deleted: ", file);
        }
    });
}, {
    runOnInit: true,
    timezone: "Asia/Kolkata"
});