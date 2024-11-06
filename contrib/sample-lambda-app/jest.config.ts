import type { Config } from "jest";

const config: Config = {
    verbose: true,
    preset: "ts-jest",
    testEnvironment: "node",
    moduleFileExtensions: ["ts", "js"],
    testRegex: "/test/.+\.ts$",
    // globalSetup: "<rootDir>/tests/jest.globalSetup.ts",
    // globalTeardown: "<rootDir>/tests/jest.globalTeardown.ts",
};

export default config;
