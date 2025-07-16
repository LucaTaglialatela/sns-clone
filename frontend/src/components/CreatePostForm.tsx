import React, { useState } from "react";
import PostForm from "./PostForm";
import { calculateSHA256 } from "../utils/hash";

interface CreatePostProps {
  onPostCreated: () => void;
}

const CreatePost: React.FC<CreatePostProps> = ({ onPostCreated }) => {
  const [text, setText] = useState<string>("");
  const [file, setFile] = useState<File | null>(null);
  const [isSubmitting, setIsSubmitting] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();

    // Remove invisible characters from the input string
    const invisibleCharRegex =
      /[\u0000-\u001F\u00A0\u115F\u1160\u2000-\u200D\u2028-\u202F\u205F\u2060\u3000\u3164\uFEFF]/g;
    const filteredText = text.replace(invisibleCharRegex, "");

    // A post can have just a file, or just text, but not be empty.
    if (!filteredText.trim() && !file) {
      return;
    }

    setIsSubmitting(true);
    setError(null);

    try {
      let image: string = "";
      // If a file has been selected, upload it to S3
      if (file) {
        const hash = await calculateSHA256(file);
        // First we get the presigned url from our backend
        const presignResponse = await fetch("/presign", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({
            fileName: file.name,
            fileType: file.type,
            fileHash: hash,
          }),
        });
        if (!presignResponse.ok) throw new Error("Could not get upload URL.");

        const { url, key } = await presignResponse.json();

        // After receiving the presigned url, we use it to upload
        // the file to S3
        const s3Response = await fetch(url, {
          method: "PUT",
          headers: { "Content-Type": file.type },
          body: file,
        });
        if (!s3Response.ok) throw new Error("Failed to upload file to S3.");

        // We set the value of image to the presigned url key
        image = key;
      }

      // After uploading to S3, we need to update our database
      const createPostResponse = await fetch("/posts", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          text,
          image,
        }),
      });

      if (!createPostResponse.ok) {
        throw new Error(
          `Error: ${createPostResponse.status} ${createPostResponse.statusText}`
        );
      }

      // After creating a post, we empty the text field and file input
      // and signal to the parent component that a post was created
      setText("");
      setFile(null);
      onPostCreated();
    } catch (err) {
      const errorMessage =
        err instanceof Error ? err.message : "An unknown error occurred.";
      setError(errorMessage);
      console.error("Failed to create post:", errorMessage);
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <PostForm
      formId={`create-post`}
      text={text}
      onTextChange={setText}
      file={file}
      onFileChange={setFile}
      isSubmitting={isSubmitting}
      onSubmit={handleSubmit}
      error={error}
      submitButtonText="Post"
    />
  );
};

export default CreatePost;
