import React, { useMemo, useRef, useState, type ChangeEvent } from "react";

interface PostFormProps {
  // Form state
  formId: string;
  text: string;
  onTextChange: (value: string) => void;
  file: File | null;
  onFileChange: (file: File | null) => void;
  fileName?: string;
  onFileNameChange?: (fileName: string) => void;

  // Submission state and logic
  isSubmitting: boolean;
  onSubmit: (event: React.FormEvent<HTMLFormElement>) => void;
  error: string | null;

  // UI Customization
  submitButtonText: string;
  placeholder?: string;
  onCancel?: () => void;
}

const PostForm: React.FC<PostFormProps> = ({
  formId,
  text,
  onTextChange,
  file,
  onFileChange,
  fileName,
  onFileNameChange,
  isSubmitting,
  onSubmit,
  error,
  submitButtonText,
  placeholder = "What's on your mind?",
  onCancel,
}) => {
  const maxLength = 280;
  const fileInputRef = useRef<HTMLInputElement>(null);
  const fileUploadId = `${formId}-file-upload`;
  const [fileError, setFileError] = useState<string>("");

  const handleFileChangeEvent = (event: ChangeEvent<HTMLInputElement>) => {
    const file =
      event.target.files && event.target.files.length > 0
        ? event.target.files[0]
        : null;

    if (!file) {
      onFileChange(null);
      return;
    }

    const maxSizeInBytes = 5 * 1024 * 1024; // 5MB
    if (file.size > maxSizeInBytes) {
      setFileError("File size exceeds the limit (5MB).");
      event.target.value = "";
      onFileChange(null);
      return;
    }

    const fileReader = new FileReader();
    fileReader.onloadend = (e) => {
      const arr = new Uint8Array(e.target?.result as ArrayBuffer).subarray(
        0,
        6
      );
      let header = "";
      for (let i = 0; i < arr.length; i++) {
        header += arr[i].toString(16);
      }

      // Check for PNG signature marker: 89 50 4E 47
      const isPng = header.startsWith("89504e47");
      // Check the first two bytes for JPEG Start of Image (SOI) marker: FF D8 FF
      const isJpg = header.startsWith("ffd8ff");
      // Check the first 6 bytes for GIF signature: GIF87a or GIF89a
      const isGif =
        header.startsWith("474946383761") || header.startsWith("474946383961");

      const isValidFileType = isPng || isJpg || isGif;
      if (!isValidFileType) {
        setFileError("Invalid file type.");
        event.target.value = "";
        onFileChange(null);
        return;
      }
    };
    fileReader.readAsArrayBuffer(file);
    setFileError("");
    onFileChange(file);
  };

  const handleRemoveFile = () => {
    onFileChange(null);
    onFileNameChange && onFileNameChange("");
    if (fileInputRef.current) {
      fileInputRef.current.value = "";
    }
  };

  const originalFileName = useMemo(() => {
    if (!fileName) return undefined;
    const firstUnderScoreIndex = fileName.indexOf("_");

    if (firstUnderScoreIndex === -1) {
      return undefined;
    }

    return fileName.substring(firstUnderScoreIndex + 1);
  }, [fileName]);

  return (
    <form onSubmit={onSubmit} className="flex flex-col gap-4">
      {error && (
        <div className="text-red-600 p-3 border border-red-300 bg-red-50 rounded-lg">
          {error}
        </div>
      )}

      <textarea
        value={text}
        onChange={(e) => {
          const invisibleCharRegex =
            /[\u0000-\u0009\u000B-\u000C\u000E-\u001F\u00A0\u115F\u1160\u2000-\u200D\u202A-\u202F\u205F\u2060\u3000\u3164\uFEFF]/g;
          const filteredText = e.target.value.replace(invisibleCharRegex, "");
          onTextChange(filteredText);
        }}
        placeholder={placeholder}
        disabled={isSubmitting}
        className="w-full p-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-0 focus:border-black focus:border-2 disabled:opacity-70"
        rows={3}
        maxLength={maxLength}
      />

      <div className="flex items-center justify-between">
        <div className="flex items-center">
          <label
            htmlFor={fileUploadId}
            className="cursor-pointer text-sm hover:text-gray-800 font-medium"
          >
            {file
              ? `Selected: ${file.name}`
              : originalFileName
              ? `Selected: ${originalFileName}`
              : "Add a file (png, jpg, jpeg, gif)."}
          </label>
          <input
            ref={fileInputRef}
            id={fileUploadId}
            name="file-upload"
            type="file"
            className="sr-only"
            onChange={handleFileChangeEvent}
            accept="image/png, image/jpeg, image/jpg, image/gif"
          />
          {(file || fileName) && (
            <button
              type="button"
              onClick={handleRemoveFile}
              className="ml-2 text-red-500 cursor-pointer hover:text-red-700 font-bold"
            >
              &times;
            </button>
          )}
          <p className="ml-3 text-red-500 text-sm font-medium">{fileError}</p>
        </div>

        <div className="flex items-center space-x-2">
          <div className="text-sm hover:text-gray-800 font-medium">
            {text.length} / {maxLength} characters
          </div>
          {onCancel && (
            <button
              type="button"
              onClick={onCancel}
              disabled={isSubmitting}
              className="px-3 py-1.5 text-sm font-semibold cursor-pointer text-gray-700 bg-gray-200 rounded-full hover:bg-gray-300 disabled:opacity-70"
            >
              Cancel
            </button>
          )}
          <button
            type="submit"
            disabled={!(text.trim() || file) || isSubmitting}
            className="cursor-pointer bg-gray-900 hover:bg-gray-700 text-white font-bold text-sm py-1.5 px-3 rounded-full focus:outline-none focus:shadow-outline disabled:bg-gray-400 disabled:cursor-not-allowed"
          >
            {isSubmitting ? "Submitting..." : submitButtonText}
          </button>
        </div>
      </div>
    </form>
  );
};

export default PostForm;
