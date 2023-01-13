
export class Utils {
    /**
     * This function decode a masked string (initially base64 encoded) and split it into comma separated items to return
     * an arry of strings
     * @param str string
     * @returns arrayt of string
     */
    static unmaskList(str: string): Array<string> {
        const serializedList = Utils.decode(str)
        return serializedList.split(",")
    }
    /**
     * This function decode a masked string (initially base64 encoded
     * @param str 
     * @returns 
     */
    static unmask(str: string): string {
        return Utils.decode(str)
    }

    /**
     * 
     * @param str Base 64 decode
     * @returns 
     */
    static decode = (str: string): string => Buffer.from(str, 'base64').toString('binary');
    /**
     * 
     * @param str Base 64 encode
     * @returns 
     */
    static encode = (str: string): string => Buffer.from(str, 'binary').toString('base64');

}
