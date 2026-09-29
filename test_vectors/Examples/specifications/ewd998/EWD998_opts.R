library(ggplot2)
library(dplyr)

data <- read.csv(header=TRUE, sep = "#", file = file.choose())

summary = summarise(group_by(data,Variant, Node),
                    mean_Length = mean(Length),
                    sd_Length = sd(Length),
                    mean_IP = mean(InitiateProbe),
                    sd_IP = sd(InitiateProbe),
                    mean_PT = mean(PassToken),
                    sd_PT = sd(PassToken),
                    mean_SM = mean(SendMsg),
                    sd_SM = sd(SendMsg),
                    mean_RM = mean(RecvMsg),
                    sd_MR = sd(RecvMsg),
                    mean_DA = mean(Deactivate),
                    sd_DA = sd(Deactivate),
                    mean_T = mean(T),
                    sd_T = sd(T),
                    mean_T2TD = mean(T2TD),
                    sd_T2TD = sd(T2TD)
)
Nodes <- unique(summary$Node)

####
#### T2TD
####
for (n in Nodes) {
  print(ggplot(filter(summary, Node == n), 
               aes(x = reorder(Variant, mean_T2TD), y = mean_T2TD, fill = Variant)) +
          geom_bar(stat = "identity") +
          geom_errorbar(aes(ymin=mean_T2TD-sd_T2TD, ymax=mean_T2TD+sd_T2TD), width=.2,
                        position=position_dodge(.9)) +
          scale_x_discrete(guide = guide_axis(n.dodge=3))+
          theme_minimal() +
          labs(
            x = "Spec variant",
            y = "Average length while terminated /\\ ~terminationDetected holds",
            title = paste(
              "Number of Nodes: ", n, "Traces:", nrow(filter(data, Node == n))
            )
          ))
}

####
#### InitiateProbe actions
####
for (n in Nodes) {
  print(ggplot(filter(summary, Node == n), 
               aes(x = reorder(Variant, mean_IP), y = mean_IP, fill = Variant)) +
          geom_bar(stat = "identity") +
          geom_errorbar(aes(ymin=mean_IP-sd_IP, ymax=mean_IP+sd_IP), width=.2,
                        position=position_dodge(.9)) +
          scale_x_discrete(guide = guide_axis(n.dodge=3))+
          theme_minimal() +
          labs(
            x = "Spec variant",
            y = "Average number of InitiateProbe actions",
            title = paste(
              "Number of Nodes: ", n, "Traces:", nrow(filter(data, Node == n))
            )
          ))
}

####
#### Length & T
####
for (n in Nodes) {
  print(ggplot(filter(summary, Node == n), 
       aes(x = reorder(Variant, mean_Length), y = mean_Length, fill = Variant)) +
  geom_bar(stat = "identity") +
  geom_errorbar(aes(ymin=mean_Length-sd_Length, ymax=mean_Length+sd_Length), width=.2,
                position=position_dodge(.9)) +
  scale_x_discrete(guide = guide_axis(n.dodge=3))+
  theme_minimal() +
  labs(
    x = "Spec variant",
    y = "Average length of behaviors",
    title = paste(
      "Number of Nodes: ", n, "Traces:", nrow(filter(data, Node == n))
    )
  ))
}

########
######## Occurrences of actions
########
for (n in Nodes) {
  print(ggplot(filter(summary, Node == n)) + 
  geom_point(aes(x=reorder(Variant, mean_PT), y = mean_PT,size=5,colour = "PassToken",shape = "PassToken")) + 
  geom_point(aes(x=reorder(Variant, mean_IP),y=mean_IP,size=5,colour = "InitiateProbe",shape = "InitiateProbe")) +
  #  geom_point(aes(x=Variant,y=mean_IP,colour = "InitiateProbe",shape = "InitiateProbe")) +
  geom_point(aes(x=reorder(Variant, mean_SM),y=mean_SM,size=5,colour = "SendMsg",shape = "SendMsg")) +
  geom_point(aes(x=reorder(Variant, mean_RM),y=mean_RM,size=5,colour = "RecvMsg",shape = "RecvMsg")) +
  geom_point(aes(x=reorder(Variant, mean_DA),y=mean_DA,size=5,colour = "Deactivate",shape = "Deactivate")) +
  ## x-axis labels should not overlap.
  scale_x_discrete(guide = guide_axis(n.dodge=3))+
  #scale_x_discrete(guide = guide_axis(check.overlap = TRUE))+
  #coord_flip() +
  theme_minimal() +
  #theme(legend.position = "none") +
  labs(
    x = "Spec variant",
    y = "Average number of occurrences in behaviors",
    title = paste(
      "Number of Nodes: ", n, " Traces:", nrow(filter(data, Node == n))
    )
  ))
}

########
######## Correlations
########

##install.packages("ggcorrplot")
library("ggcorrplot")
my_data <- filter(summary, Node == 113)[, c("mean_Length", "mean_SM", "mean_RM", "mean_IP", "mean_PT", "mean_DA", "mean_T")]
p.mat <- cor_pmat(my_data)
## Check for correlation in 'data'
## 'spearman' (3) correlation because data has no normal distribution
## (see previous plots).
corr <- round(cor(my_data), 3)
ggcorrplot(corr, p.mat = cor_pmat(my_data),
           hc.order = TRUE, type = "lower",
           color = c("#FC4E07", "white", "#00AFBB"),
           outline.col = "white", lab = TRUE)
