class Solution {
    public String removeKdigits(String num, int k) {
        if (k == num.length()) return "0";
        StringBuilder sb = new StringBuilder();
        for (char c : num.toCharArray()) {
            while (k > 0 && sb.length() > 0 && sb.charAt(sb.length() - 1) > c) {
                sb.deleteCharAt(sb.length() - 1);
                k--;
            }

            sb.append(c);
        }
        while (k > 0) {
            sb.deleteCharAt(sb.length() - 1);
            k--;
        }
        String result = sb.toString();
        result = result.replaceAll("^0+", "");

        return result.isEmpty() ? "0" : result;
        
    }
}